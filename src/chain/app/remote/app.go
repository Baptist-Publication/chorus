package remote

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/big"
	"sync"

	ethcmn "github.com/Baptist-Publication/chorus-module/lib/eth/common"
	ethtypes "github.com/Baptist-Publication/chorus-module/lib/eth/core/types"
	"github.com/Baptist-Publication/chorus-module/lib/eth/rlp"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/merkle"
	"go.uber.org/zap"

	pbtypes "github.com/Baptist-Publication/angine/protos/types"
	agtypes "github.com/Baptist-Publication/angine/types"
	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
	crypto "github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	civil "github.com/Baptist-Publication/chorus/src/chain/node"
	cvstate "github.com/Baptist-Publication/chorus/src/tools/state"
	"github.com/tendermint/tmlibs/db"
)

const (
	ReceiptsPrefix      = "receipts-"
	OfficialAddress     = "0x7752b42608a0f1943c19fc5802cb027e60b4c911"
	StateRemoveEmptyObj = true
)

var (
	RemoteTxTag = []byte{'r', 'm', 't', 0x01}
)

type LastBlockInfo struct {
	Height  agtypes.INT
	AppHash []byte
}

type StoreReceipt struct {
	txHash  []byte
	receipt []byte
}

type stateDup struct {
	height     agtypes.INT
	round      agtypes.INT
	key        string
	state      *cvstate.RemoteState
	lock       *sync.Mutex
	execFinish chan agtypes.ExecuteResult
	quit       chan struct{}
	receipts   []StoreReceipt
}

type RemoteApp struct {
	agtypes.BaseApplication
	agtypes.CommApplication

	core civil.Core

	dbDir string

	logger       *zap.Logger
	stateMtx     sync.Mutex // protected concurrent changes of app.state
	state        *cvstate.RemoteState
	chainDb      db.DB // Block chain database
	privkey      crypto.PrivKeyEd25519
	stateDupsMtx sync.RWMutex // protect concurrent changes of app fields
	stateDups    map[string]*stateDup

	attributes map[string]string

	Config      *viper.Viper
	AngineHooks agtypes.Hooks

	remote *RemoteAppClient
}

var (
	EthSigner   = ethtypes.HomesteadSigner{}
	IsHomestead = true

	lastBlockKey = []byte("lastblock")
	big0         = big.NewInt(0)

	errQuitExecute = fmt.Errorf("quit executing block")
)

func init() {
	if _, ok := civil.Apps["remote"]; ok {
		cmn.PanicSanity("app name is preoccupied")
	}
	civil.Apps["remote"] = func(l *zap.Logger, c *viper.Viper, p crypto.PrivKey) (civil.Application, error) {
		return NewRemoteApp(l, c, p)
	}
}

func newStateDup(logger *zap.Logger, state *cvstate.RemoteState, block *agtypes.BlockCache, height, round agtypes.INT) *stateDup {
	stateCopy := state.Copy()
	if stateCopy == nil {
		logger.Error("state deep copy failed")
		return nil
	}
	return &stateDup{
		height:     height,
		round:      round,
		key:        stateKey(block, height, round),
		state:      stateCopy,
		lock:       &sync.Mutex{},
		quit:       make(chan struct{}, 1),
		execFinish: make(chan agtypes.ExecuteResult, 1),
	}
}

func stateKey(block *agtypes.BlockCache, height, round agtypes.INT) string {
	return ethcmn.Bytes2Hex(block.Hash())
}

func NewRemoteApp(logger *zap.Logger, config *viper.Viper, privkey crypto.PrivKey) (*RemoteApp, error) {
	serverAddr := config.GetString("rpcapp_addr")
	if len(serverAddr) == 0 {
		logger.Warn("rpcapp_addr config not found")
		return nil, APP_ERR
	}

	var remote RemoteAppClient
	if !remote.Init(logger, config) {
		return nil, APP_ERR
	}

	app := &RemoteApp{
		dbDir:      config.GetString("db_dir"),
		stateDups:  make(map[string]*stateDup),
		logger:     logger,
		privkey:    *(privkey.(*crypto.PrivKeyEd25519)),
		attributes: make(map[string]string),
		remote:     &remote,
		Config:     config,
	}

	app.AngineHooks = agtypes.Hooks{
		OnNewRound: agtypes.NewHook(app.OnNewRound),
		OnCommit:   agtypes.NewHook(app.OnCommit),
		// OnPrevote:  types.NewHook(app.OnPrevote),
		OnExecute: agtypes.NewHook(app.OnExecute),
	}

	var err error
	var stateDB db.DB
	if stateDB, err = db.NewGoLevelDB("remoteStatus", app.dbDir); err != nil {
		return nil, err
	}
	app.state = cvstate.NewRemoteState(stateDB, logger)

	if err = app.BaseApplication.InitBaseApplication("remotestate", app.dbDir); err != nil {
		app.logger.Error("InitBaseApplication error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}
	if app.chainDb, err = db.NewGoLevelDB("remotechaindata", app.dbDir); err != nil {
		app.logger.Error("open go level DB error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	return app, nil
}

func (app *RemoteApp) SetCore(core civil.Core) {
	app.core = core
}

func (app *RemoteApp) Start() (err error) {
	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	app.state.Load(lastBlock.AppHash)
	app.remote.OnStart()

	fmt.Printf("remote_app is running\n")
	app.logger.Info("app started")
	return nil
}

func (app *RemoteApp) Stop() {
	app.BaseApplication.Stop()
	app.chainDb.Close()
	app.remote.OnStop()
}

func (app *RemoteApp) GetAttributes() map[string]string {
	return app.attributes
}

func (app *RemoteApp) GetAngineHooks() agtypes.Hooks {
	return app.AngineHooks
}

func (app *RemoteApp) CompatibleWithAngine() {}

// ExecuteRemoteTx execute tx one by one in the loop, without lock, so should always be called between Lock() and Unlock() on the *stateDup
func (app *RemoteApp) ExecuteRemoteTx(stateDup *stateDup, blockHash ethcmn.Hash, bs []byte, txIndex int) (hash []byte, err error) {
	txBytes := agtypes.UnwrapTx(bs)
	tx := new(ethtypes.Transaction)
	stateSnapshot := stateDup.state.Snapshot()
	if err = rlp.DecodeBytes(txBytes, tx); err != nil {
		return
	}

	var sender ethcmn.Address
	sender, err = ethtypes.Sender(EthSigner, tx)
	if err != nil {
		return
	}

	stateDup.state.AddNonce(sender.Hex(), 1)

	var receipt []byte
	receipt, err = app.remote.ApplyTransaction(stateDup, sender, tx.Data())
	if err != nil {
		stateDup.state.RevertToSnapshot(stateSnapshot)
		return
	}
	txHash := tx.Hash().Bytes()
	if len(receipt) != 0 {
		storeR := StoreReceipt{
			txHash:  txHash,
			receipt: receipt,
		}
		stateDup.receipts = append(stateDup.receipts, storeR)
	}

	return tx.Hash().Bytes(), err
}

func (app *RemoteApp) OnNewRound(height, round agtypes.INT, block *agtypes.BlockCache) (interface{}, error) {
	return agtypes.NewRoundResult{}, nil
}

func (app *RemoteApp) OnPrevote(height, round agtypes.INT, block *agtypes.BlockCache) (interface{}, error) {
	if block == nil {
		return nil, nil
	}
	sk := stateKey(block, height, round)

	app.stateDupsMtx.Lock()
	if _, ok := app.stateDups[sk]; ok {
		app.stateDupsMtx.Unlock()
		return nil, nil
	}
	app.stateMtx.Lock()
	stateDup := newStateDup(app.logger, app.state, block, height, round)
	app.stateMtx.Unlock()
	app.stateDups[sk] = stateDup
	app.stateDupsMtx.Unlock()

	stateDup.lock.Lock()
	execRes := agtypes.ExecuteResult{}
	defer func() {
		stateDup.execFinish <- execRes
		stateDup.lock.Unlock()
	}()

	if block.Data == nil || len(block.Data.Txs) == 0 {
		return nil, nil
	}

	blockHash := ethcmn.BytesToHash(block.Hash())
	for i, tx := range block.Data.Txs {
		select {
		case <-stateDup.quit:
			// log quit, caused by failed consensus or ...
			execRes.Error = errQuitExecute
			return nil, errQuitExecute
		default:
			// we only care about evm txs here
			if !bytes.HasPrefix(tx, RemoteTxTag) {
				continue
			}

			if txHash, err := app.ExecuteRemoteTx(stateDup, blockHash, tx, i); err != nil {
				execRes.InvalidTxs = append(execRes.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: txHash, Error: err})
			} else {
				execRes.ValidTxs = append(execRes.ValidTxs, tx)
			}
		}
	}

	return nil, nil
}

func (app *RemoteApp) OnExecute(height, round agtypes.INT, block *agtypes.BlockCache) (interface{}, error) {
	var (
		res agtypes.ExecuteResult
		err error

		sk = stateKey(block, height, round)

		eventData = make([][]byte, 0)
	)
	// normal transaction
	app.stateDupsMtx.Lock()
	if st, ok := app.stateDups[sk]; ok {
		res = <-st.execFinish
	} else {
		app.stateMtx.Lock()
		stateDup := newStateDup(app.logger, app.state, block, height, round)
		app.stateMtx.Unlock()

		blockHash := ethcmn.BytesToHash(block.Hash())

		stateDup.lock.Lock()
		for i, tx := range block.Data.Txs {
			if !bytes.Equal(tx[:4], RemoteTxTag) {
				continue
			}

			if txHash, err := app.ExecuteRemoteTx(stateDup, blockHash, tx, i); err != nil {
				res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: txHash, Error: err})
			} else {
				res.ValidTxs = append(res.ValidTxs, tx)
				eventData = append(eventData, tx[:])
			}
		}
		stateDup.lock.Unlock()

		app.stateDups[sk] = stateDup
	}
	app.stateDupsMtx.Unlock()

	return res, err
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateMtx is still needed
func (app *RemoteApp) OnCommit(height, round agtypes.INT, block *agtypes.BlockCache) (interface{}, error) {
	var (
		appHash []byte
		err     error

		sk = stateKey(block, height, round)
	)

	if _, ok := app.stateDups[sk]; !ok {
		app.stateMtx.Lock()
		appHash, err = app.state.Commit()
		if err != nil {
			// TODO
		}
		app.stateMtx.Unlock()
		app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash})
		return agtypes.CommitResult{AppHash: appHash}, nil
	}

	app.stateDups[sk].lock.Lock()
	appHash, err = app.stateDups[sk].state.Commit()
	app.stateDups[sk].lock.Unlock()

	if err != nil {
		app.stateMtx.Lock()
		appHash, err = app.state.Commit()
		if err != nil {
			return nil, err
		}
		app.stateMtx.Unlock()
		app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash})
		return nil, err
	}

	app.stateMtx.Lock()
	app.state.Reload(appHash)
	app.stateMtx.Unlock()
	app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash})
	rHash := app.SaveReceipts(app.stateDups[sk])
	delete(app.stateDups, sk)

	return agtypes.CommitResult{
		AppHash:      appHash,
		ReceiptsHash: rHash,
	}, nil
}

func (app *RemoteApp) IncomingEvent(from string, height agtypes.INT) bool {
	return true
}

func (app *RemoteApp) HandleEvent(eventData []byte, notification *civil.EventNotificationTx) {
	buf := bytes.NewBuffer(eventData)
	dec := gob.NewDecoder(buf)
	var data [][]byte
	if err := dec.Decode(&data); err != nil {
		app.logger.Error("fail to decode event", zap.Error(err))
		return
	}

	for _, tx := range data {
		if !bytes.Equal(tx[:4], RemoteTxTag) {
			continue
		}
		etx := new(ethtypes.Transaction)
		if err := rlp.DecodeBytes(tx[4:], etx); err != nil {
			fmt.Println(err)
		}
		app.logger.Debug("event", zap.String("ethereum transaction", fmt.Sprintf("%+v", etx)))
	}
}

func (app *RemoteApp) SaveReceipts(stdup *stateDup) []byte {
	savedReceipts := make([][]byte, 0, len(stdup.receipts))
	receiptBatch := app.chainDb.NewBatch()

	for i := range stdup.receipts {
		receipt := stdup.receipts[i]
		key := append([]byte(ReceiptsPrefix), receipt.txHash...)
		receiptBatch.Set(key, receipt.receipt)
		savedReceipts = append(savedReceipts, receipt.txHash)
	}
	receiptBatch.Write()

	return merkle.SimpleHashFromHashes(savedReceipts)
}

func (app *RemoteApp) CheckTx(bs []byte) error {
	txBytes := agtypes.UnwrapTx(bs)

	if bytes.HasPrefix(bs, RemoteTxTag) {
		var err error
		tx := new(ethtypes.Transaction)
		err = rlp.DecodeBytes(txBytes, tx)
		if err != nil {
			return err
		}
		from, _ := ethtypes.Sender(EthSigner, tx)
		var nonce uint64
		var balance *big.Int
		app.stateMtx.Lock()
		fromHex := from.Hex()
		nonce = app.state.GetNonce(fromHex)
		balance = app.state.GetBalance(fromHex)
		app.stateMtx.Unlock()
		if nonce > tx.Nonce() {
			return fmt.Errorf("nonce too low")
		}
		if balance == nil || balance.Cmp(tx.Cost()) < 0 {
			return fmt.Errorf("not enough funds")
		}
		app.remote.CheckTx(from, tx.Data())
		return nil
	}

	return fmt.Errorf("unknown transaction")
}

func (app *RemoteApp) Info() (resInfo agtypes.ResultInfo) {
	lb := &LastBlockInfo{
		AppHash: make([]byte, 0),
		Height:  0,
	}
	if res, err := app.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}
	resInfo.LastBlockAppHash = lb.AppHash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "alpha 0.2"
	resInfo.Data = "for remote invoke"
	return
}

func (app *RemoteApp) Query(query []byte) agtypes.Result {
	if len(query) == 0 {
		return agtypes.NewResultOK([]byte{}, "Empty query")
	}
	var res agtypes.Result
	action := query[0]
	load := query[1:]
	switch action {
	case 0:
		res = app.queryContract(load)
	case 1:
		res = app.queryNonce(load)
	case 2:
		res = app.queryBalance(load)
	case 3:
		res = app.queryReceipt(load)
		//	case 4:
		//		res = app.queryContractExistence(load)
	default:
		res = agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "unimplemented query")
	}

	// check if contract exists
	return res
}

//func (app *RemoteApp) queryContractExistence(load []byte) types.Result {
//	tx := new(ethtypes.Transaction)
//	err := rlp.DecodeBytes(load, tx)
//	if err != nil {
//		// logger.Error("fail to decode tx: ", err)
//		return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
//	}
//	contractAddr := tx.To()
//
//	app.stateMtx.Lock()
//	hashBytes := app.state.GetCodeHash(*contractAddr).Bytes()
//	app.stateMtx.Unlock()
//
//	if bytes.Equal(tx.Data(), hashBytes) {
//		return types.NewResultOK(append([]byte{}, byte(0x01)), "contract exists")
//	}
//	return types.NewResultOK(append([]byte{}, byte(0x00)), "constract doesn't exist")
//}

func (app *RemoteApp) queryContract(load []byte) agtypes.Result {
	tx := new(ethtypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		// logger.Error("fail to decode tx: ", err)
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, err.Error())
	}

	txMsg, err := tx.AsMessage(EthSigner)
	var res []byte
	if err == nil {
		app.stateMtx.Lock()
		stateDup := newStateDup(app.logger, app.state, nil, 0, 0)
		res, err = app.remote.ApplyTransaction(stateDup, txMsg.From(), txMsg.Data())
		app.stateMtx.Unlock()
	}
	if err != nil {
		return agtypes.NewError(pbtypes.CodeType_InternalError, err.Error())
	}

	return agtypes.NewResultOK(res, "")
}

func (app *RemoteApp) queryNonce(addrBytes []byte) agtypes.Result {
	if len(addrBytes) != 20 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes).Hex()

	app.stateMtx.Lock()
	nonce := app.state.GetNonce(addr)
	app.stateMtx.Unlock()

	data, err := rlp.EncodeToBytes(nonce)
	if err != nil {
		// logger.Warn("query error", err)
	}
	return agtypes.NewResultOK(data, "")
}

func (app *RemoteApp) queryBalance(addrBytes []byte) agtypes.Result {
	if len(addrBytes) != 20 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes).Hex()

	app.stateMtx.Lock()
	balance := app.state.GetBalance(addr)
	app.stateMtx.Unlock()

	data, err := rlp.EncodeToBytes(balance)
	if err != nil {
		// logger.Warn("query error", err)
	}
	return agtypes.NewResultOK(data, "")
}

func (app *RemoteApp) queryReceipt(txHashBytes []byte) agtypes.Result {
	if len(txHashBytes) != 32 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid txhash")
	}
	key := append([]byte(ReceiptsPrefix), txHashBytes...)
	data := app.chainDb.Get(key)
	if len(data) == 0 {
		return agtypes.NewError(pbtypes.CodeType_InternalError, "fail to get receipt for tx:"+string(key))
	}
	return agtypes.NewResultOK(data, "")
}

func (app *RemoteApp) amProposer() bool {
	engine := app.core.GetEngine()
	_, vals := engine.GetValidators()
	return engine.PrivValidator().GetPubKey().Equals(vals.Proposer().GetPubKey())
}
