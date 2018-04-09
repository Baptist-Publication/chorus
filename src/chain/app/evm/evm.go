package evm

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"time"

	pbtypes "github.com/Baptist-Publication/angine/protos/types"
	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-merkle"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
	ethcmn "github.com/Baptist-Publication/chorus/src/eth/common"
	ethcore "github.com/Baptist-Publication/chorus/src/eth/core"
	ethstate "github.com/Baptist-Publication/chorus/src/eth/core/state"
	ethtypes "github.com/Baptist-Publication/chorus/src/eth/core/types"
	ethvm "github.com/Baptist-Publication/chorus/src/eth/core/vm"
	ethcrypto "github.com/Baptist-Publication/chorus/src/eth/crypto"
	ethdb "github.com/Baptist-Publication/chorus/src/eth/ethdb"
	ethparams "github.com/Baptist-Publication/chorus/src/eth/params"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	OfficialAddress     = "0x7752b42608a0f1943c19fc5802cb027e60b4c911"
	StateRemoveEmptyObj = true
	APP_NAME            = "evm"

	DatabaseCache   = 128
	DatabaseHandles = 1024
)

var (
	ReceiptsPrefix = []byte("receipts-")
	ABIPrefix      = []byte("solidity-abi-")

	EVMTag                 = []byte{'e', 'v', 'm'}
	EVMTxTag               = append(EVMTag, 0x01)
	EVMCreateContractTxTag = append(EVMTag, 0x02)
)

// CreateContractTx wraps ethereum tx bytes with the abi json bytes for this contract
type CreateContractTx struct {
	EthTx  []byte
	EthAbi []byte
}

type LastBlockInfo struct {
	Height  def.INT
	AppHash []byte
}

type stateDup struct {
	height     def.INT
	round      def.INT
	key        string
	state      *ethstate.StateDB
	lock       *sync.Mutex
	execFinish chan agtypes.ExecuteResult
	quit       chan struct{}
	receipts   ethtypes.Receipts
}

type abiBox = struct {
	key []byte
	val []byte
}

type EVMApp struct {
	agtypes.BaseApplication

	datadir string

	logger        *zap.Logger
	stateMtx      sync.Mutex // protected concurrent changes of app.state
	state         *ethstate.StateDB
	currentHeader *ethtypes.Header
	chainConfig   *ethparams.ChainConfig
	chainDb       ethdb.Database // Block chain database
	blockChain    *ethcore.BlockChain
	stateDupsMtx  sync.RWMutex // protect concurrent changes of app fields
	stateDups     map[string]*stateDup

	abis []abiBox

	Config      *viper.Viper
	AngineHooks agtypes.Hooks
}

var (
	EmptyTrieRoot = ethcmn.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	EthSigner     = ethtypes.HomesteadSigner{}
	IsHomestead   = true

	lastBlockKey = []byte("lastblock")
	evmConfig    = ethvm.Config{DisableGasMetering: false, EnableJit: true, ForceJit: true}
	big0         = big.NewInt(0)

	errQuitExecute = fmt.Errorf("quit executing block")
)

func makeCurrentHeader(block *agtypes.BlockCache) *ethtypes.Header {
	return &ethtypes.Header{
		ParentHash: ethcmn.HexToHash("0x00"),
		Difficulty: big0,
		GasLimit:   ethcmn.MaxBig,
		Number:     ethparams.MainNetSpuriousDragon,
		Time:       big.NewInt(block.Header.Time),
	}
}

func newStateDup(logger *zap.Logger, state *ethstate.StateDB, block *agtypes.BlockCache, height, round def.INT) *stateDup {
	stateCopy := state.DeepCopy()
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

func newABIs() []abiBox {
	return make([]abiBox, 0, 256)
}

func stateKey(block *agtypes.BlockCache, height, round def.INT) string {
	return ethcmn.Bytes2Hex(block.Hash())
}

func OpenDatabase(datadir string, name string, cache int, handles int) (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(filepath.Join(datadir, name), cache, handles)
}

func NewEVMApp(logger *zap.Logger, config *viper.Viper /*, privkey crypto.PrivKey*/) (*EVMApp, error) {
	app := &EVMApp{
		datadir:     config.GetString("db_dir"),
		chainConfig: new(ethparams.ChainConfig),
		stateDups:   make(map[string]*stateDup),
		logger:      logger,

		abis: newABIs(),

		Config: config,
	}

	app.AngineHooks = agtypes.Hooks{
		OnNewRound: agtypes.NewHook(app.OnNewRound),
		OnCommit:   agtypes.NewHook(app.OnCommit),
		// OnPrevote:  agtypes.NewHook(app.OnPrevote),
		OnExecute: agtypes.NewHook(app.OnExecute),
	}

	var err error
	if err = app.BaseApplication.InitBaseApplication(APP_NAME, app.datadir); err != nil {
		app.logger.Error("InitBaseApplication error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}
	if app.chainDb, err = OpenDatabase(app.datadir, "chaindata", DatabaseCache, DatabaseHandles); err != nil {
		app.logger.Error("OpenDatabase error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	return app, nil
}

func (app *EVMApp) Start() (err error) {
	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if err != nil {
		app.logger.Error("fail to load last block", zap.Error(err))
		return
	}

	trieRoot := EmptyTrieRoot
	if len(lastBlock.AppHash) > 0 {
		trieRoot = ethcmn.BytesToHash(lastBlock.AppHash)
	}
	if app.state, err = ethstate.New(trieRoot, app.chainDb); err != nil {
		app.Stop()
		app.logger.Error("fail to new ethstate", zap.Error(err))
		return
	}

	return nil
}

func (app *EVMApp) Stop() {
	app.BaseApplication.Stop()
	//app.EventAppBase.Stop()

	app.chainDb.Close()
}

func (app *EVMApp) GetAngineHooks() agtypes.Hooks {
	return app.AngineHooks
}

func (app *EVMApp) CompatibleWithAngine() {}

// ExecuteEVMTx execute tx one by one in the loop, without lock, so should always be called between Lock() and Unlock() on the *stateDup
func (app *EVMApp) ExecuteEVMTx(stateDup *stateDup, header *ethtypes.Header, blockHash ethcmn.Hash, bs []byte, txIndex int) (hash []byte, err error) {
	state := stateDup.state
	stateSnapshot := state.Snapshot()
	txBytes := agtypes.UnwrapTx(bs)
	tx := new(ethtypes.Transaction)
	if err = rlp.DecodeBytes(txBytes, tx); err != nil {
		return
	}

	gp := new(ethcore.GasPool).AddGas(ethcmn.MaxBig)
	state.StartRecord(tx.Hash(), blockHash, txIndex)
	receipt, _, err := ethcore.ApplyTransaction(
		app.chainConfig,
		nil,
		gp,
		state,
		header,
		tx,
		big0,
		evmConfig)

	if err != nil {
		state.RevertToSnapshot(stateSnapshot)
		return
	}

	if receipt != nil {
		stateDup.receipts = append(stateDup.receipts, receipt)
	}

	return tx.Hash().Bytes(), err
}

func (app *EVMApp) OnNewRound(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	return agtypes.NewRoundResult{}, nil
}

func (app *EVMApp) OnPrevote(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
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
	currentHeader := makeCurrentHeader(block)
	for i, tx := range block.Data.Txs {
		select {
		case <-stateDup.quit:
			// log quit, caused by failed consensus or ...
			execRes.Error = errQuitExecute
			return nil, errQuitExecute
		default:
			// we only care about evm txs here
			if !bytes.HasPrefix(tx, EVMTxTag) {
				continue
			}

			if txHash, err := app.ExecuteEVMTx(stateDup, currentHeader, blockHash, tx, i); err != nil {
				execRes.InvalidTxs = append(execRes.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: txHash, Error: err})
			} else {
				execRes.ValidTxs = append(execRes.ValidTxs, txHash)
			}
		}
	}

	return nil, nil
}

func (app *EVMApp) OnExecute(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	var (
		res agtypes.ExecuteResult
		err error

		sk = stateKey(block, height, round)
	)

	// normal transaction
	app.stateDupsMtx.Lock()
	if st, ok := app.stateDups[sk]; ok {
		res = <-st.execFinish
	} else {
		app.stateMtx.Lock()
		stateDup := newStateDup(app.logger, app.state, block, height, round)
		app.stateMtx.Unlock()

		currentHeader := makeCurrentHeader(block)
		blockHash := ethcmn.BytesToHash(block.Hash())

		stateDup.lock.Lock()
		for i, tx := range block.Data.Txs {
			txType := tx[:4]
			switch {
			case bytes.Equal(txType, EVMTxTag):
				// txhash, err
				if _, err := app.ExecuteEVMTx(stateDup, currentHeader, blockHash, tx, i); err != nil {
					res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: tx, Error: err})
				}
			case bytes.Equal(txType, EVMCreateContractTxTag):
				txCreate, err := DecodeCreateContract(agtypes.UnwrapTx(tx))
				if err != nil {
					res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: tx, Error: err})
					continue
				}
				if _, err := app.ExecuteEVMTx(stateDup, currentHeader, blockHash, agtypes.WrapTx(EVMTxTag, txCreate.EthTx), i); err != nil {
					res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: tx, Error: err})
				} else {
					res.ValidTxs = append(res.ValidTxs, tx)

					etx := new(ethtypes.Transaction)
					// here we know this can't fail
					rlp.DecodeBytes(txCreate.EthTx, etx)
					sender, _ := ethtypes.Sender(EthSigner, etx)
					createdAddress := ethcrypto.CreateAddress(sender, etx.Nonce())
					app.abis = append(app.abis, abiBox{
						key: append(ABIPrefix, createdAddress.Bytes()...),
						val: txCreate.EthAbi,
					})
				}

			}

		}
		stateDup.lock.Unlock()

		app.stateDups[sk] = stateDup
	}
	app.stateDupsMtx.Unlock()

	return res, err
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateMtx is still needed
func (app *EVMApp) OnCommit(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	var (
		appHash ethcmn.Hash
		err     error

		sk = stateKey(block, height, round)
	)

	if _, ok := app.stateDups[sk]; !ok {
		app.stateMtx.Lock()
		appHash = app.state.IntermediateRoot(StateRemoveEmptyObj)
		app.stateMtx.Unlock()
		app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash.Bytes()})
		return agtypes.CommitResult{AppHash: appHash.Bytes()}, nil
	}

	app.stateDups[sk].lock.Lock()
	appHash, err = app.stateDups[sk].state.Commit(StateRemoveEmptyObj)
	app.stateDups[sk].lock.Unlock()

	if err != nil {
		app.stateMtx.Lock()
		appHash = app.state.IntermediateRoot(StateRemoveEmptyObj)
		app.stateMtx.Unlock()
		app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash.Bytes()})
		return nil, err
	}

	app.stateMtx.Lock()
	app.state, _ = app.stateDups[sk].state.New(appHash)
	app.stateMtx.Unlock()
	app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash.Bytes()})
	rHash := app.SaveReceipts(app.stateDups[sk])
	delete(app.stateDups, sk)

	// ignore: abis hash & error
	if len(app.abis) > 0 {
		if _, err := app.saveABIs(); err != nil {
			app.logger.Error("[saveABIs]", zap.Error(err))
		}
	}

	return agtypes.CommitResult{
		AppHash:      appHash.Bytes(),
		ReceiptsHash: rHash,
	}, nil
}

func (app *EVMApp) CheckTx(bs []byte) error {
	var err error
	txBytes := agtypes.UnwrapTx(bs)

	if bytes.HasPrefix(bs, EVMCreateContractTxTag) {
		cctx, err := DecodeCreateContract(txBytes)
		if err != nil {
			return errors.Wrap(err, "[EVMApp CheckTx]")
		}
		tx := new(ethtypes.Transaction)
		if err = rlp.DecodeBytes(cctx.EthTx, tx); err != nil {
			return errors.Wrap(err, "[EVMApp CheckTx]")
		}
		from, _ := ethtypes.Sender(EthSigner, tx)
		app.stateMtx.Lock()
		defer app.stateMtx.Unlock()
		if app.state.GetNonce(from) > tx.Nonce() {
			return fmt.Errorf("nonce too low")
		}
		if app.state.GetBalance(from).Cmp(tx.Cost()) < 0 {
			return fmt.Errorf("not enough funds")
		}
		return nil
	} else if bytes.HasPrefix(bs, EVMTxTag) {
		tx := new(ethtypes.Transaction)
		err = rlp.DecodeBytes(txBytes, tx)
		if err != nil {
			return err
		}
		from, _ := ethtypes.Sender(EthSigner, tx)
		app.stateMtx.Lock()
		defer app.stateMtx.Unlock()
		if app.state.GetNonce(from) > tx.Nonce() {
			return fmt.Errorf("nonce too low")
		}
		if app.state.GetBalance(from).Cmp(tx.Cost()) < 0 {
			return fmt.Errorf("not enough funds")
		}
		return nil
	}

	return nil
}

func (app *EVMApp) saveABIs() ([]byte, error) {
	batch := app.chainDb.NewBatch()
	for i := range app.abis {
		batch.Put(app.abis[i].key, app.abis[i].val)
	}

	if err := batch.Write(); err != nil {
		return nil, errors.Wrap(err, "[EVMApp saveABIs]")
	}

	rh := merkle.SimpleHashFromBinary(app.abis)
	app.abis = newABIs()

	return rh, nil
}

func (app *EVMApp) SaveReceipts(stdup *stateDup) []byte {
	savedReceipts := make([][]byte, 0, len(stdup.receipts))
	receiptBatch := app.chainDb.NewBatch()

	for _, receipt := range stdup.receipts {
		storageReceipt := (*ethtypes.ReceiptForStorage)(receipt)
		storageReceiptBytes, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			app.logger.Error("wrong rlp encode", zap.Error(err))
			continue
		}
		key := append(ReceiptsPrefix, receipt.TxHash.Bytes()...)
		if err := receiptBatch.Put(key, storageReceiptBytes); err != nil {
			app.logger.Error("batch receipt failed", zap.Error(err))
			continue
		}
		savedReceipts = append(savedReceipts, storageReceiptBytes)
	}
	if err := receiptBatch.Write(); err != nil {
		app.logger.Error("persist receipts failed", zap.Error(err))
	}

	return merkle.SimpleHashFromHashes(savedReceipts)
}

func (app *EVMApp) Info() (resInfo agtypes.ResultInfo) {
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
	resInfo.Data = "evm-1.5.9 with cosi and eventtx"
	return
}

func (app *EVMApp) Query(query []byte) agtypes.Result {
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
	case 4:
		res = app.queryContractExistence(load)
	default:
		res = agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "unimplemented query")
	}

	// check if contract exists
	return res
}

func (app *EVMApp) queryContractExistence(load []byte) agtypes.Result {
	tx := new(ethtypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		// logger.Error("fail to decode tx: ", err)
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, err.Error())
	}
	contractAddr := tx.To()

	app.stateMtx.Lock()
	hashBytes := app.state.GetCodeHash(*contractAddr).Bytes()
	app.stateMtx.Unlock()

	if bytes.Equal(tx.Data(), hashBytes) {
		return agtypes.NewResultOK(append([]byte{}, byte(0x01)), "contract exists")
	}
	return agtypes.NewResultOK(append([]byte{}, byte(0x00)), "constract doesn't exist")
}

func (app *EVMApp) queryContract(load []byte) agtypes.Result {
	tx := new(ethtypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		// logger.Error("fail to decode tx: ", err)
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, err.Error())
	}

	fakeHeader := &ethtypes.Header{
		ParentHash: ethcmn.HexToHash("0x00"),
		Difficulty: big0,
		GasLimit:   ethcmn.MaxBig,
		Number:     ethparams.MainNetSpuriousDragon,
		Time:       big.NewInt(time.Now().Unix()),
	}
	txMsg, _ := tx.AsMessage(EthSigner)
	envCxt := ethcore.NewEVMContext(txMsg, fakeHeader, nil)

	app.stateMtx.Lock()
	vmEnv := ethvm.NewEVM(envCxt, app.state.Copy(), app.chainConfig, evmConfig)
	gpl := new(ethcore.GasPool).AddGas(ethcmn.MaxBig)
	res, _, err := ethcore.ApplyMessage(vmEnv, txMsg, gpl) // we don't care about gasUsed
	if err != nil {
		// logger.Debug("transition error", err)
	}
	app.stateMtx.Unlock()

	return agtypes.NewResultOK(res, "")
}

func (app *EVMApp) queryNonce(addrBytes []byte) agtypes.Result {
	if len(addrBytes) != 20 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.stateMtx.Lock()
	nonce := app.state.GetNonce(addr)
	app.stateMtx.Unlock()

	data, err := rlp.EncodeToBytes(nonce)
	if err != nil {
		// logger.Warn("query error", err)
	}
	return agtypes.NewResultOK(data, "")
}

func (app *EVMApp) queryBalance(addrBytes []byte) agtypes.Result {
	if len(addrBytes) != 20 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.stateMtx.Lock()
	balance := app.state.GetBalance(addr)
	app.stateMtx.Unlock()

	data, err := rlp.EncodeToBytes(balance)
	if err != nil {
		// logger.Warn("query error", err)
	}
	return agtypes.NewResultOK(data, "")
}

func (app *EVMApp) queryReceipt(txHashBytes []byte) agtypes.Result {
	if len(txHashBytes) != 32 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid txhash")
	}
	key := append(ReceiptsPrefix, txHashBytes...)
	data, err := app.chainDb.Get(key)
	if err != nil {
		return agtypes.NewError(pbtypes.CodeType_InternalError, "fail to get receipt for tx:"+string(key))
	}
	return agtypes.NewResultOK(data, "")
}

func EncodeCreateContract(tx CreateContractTx) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(tx); err != nil {
		return nil, errors.Wrap(err, "[EncodeCreateContract]")
	}

	return buf.Bytes(), nil
}

func DecodeCreateContract(bs []byte) (*CreateContractTx, error) {
	tx := new(CreateContractTx)
	if err := gob.NewDecoder(bytes.NewReader(bs)).Decode(tx); err != nil {
		return nil, errors.Wrap(err, "[DecodeCreateContract]")
	}
	return tx, nil
}
