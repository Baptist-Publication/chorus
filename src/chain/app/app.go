package app

import (
	"bytes"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"

	"github.com/Baptist-Publication/angine"
	agtypes "github.com/Baptist-Publication/angine/types"
	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
	db "github.com/Baptist-Publication/chorus-module/lib/go-db"
	"github.com/Baptist-Publication/chorus-module/lib/go-merkle"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
	appconfig "github.com/Baptist-Publication/chorus/src/chain/config"
	ethcmn "github.com/Baptist-Publication/chorus/src/eth/common"
	ethcore "github.com/Baptist-Publication/chorus/src/eth/core"
	ethstate "github.com/Baptist-Publication/chorus/src/eth/core/state"
	ethtypes "github.com/Baptist-Publication/chorus/src/eth/core/types"
	ethvm "github.com/Baptist-Publication/chorus/src/eth/core/vm"
	ethdb "github.com/Baptist-Publication/chorus/src/eth/ethdb"
	ethparams "github.com/Baptist-Publication/chorus/src/eth/params"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
	"github.com/Baptist-Publication/chorus/src/types"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/zap"
)

const (
	StateRemoveEmptyObj = true
	APP_NAME            = "evm"
	lastBlockKey        = "lastblock"
	stateRootPrefix     = "state-prefix-"

	DatabaseCache   = 128
	DatabaseHandles = 1024
)

type LastBlockInfo struct {
	Height         def.INT
	EvmStateHash   []byte
	ShareStateHash []byte
}

func (lb *LastBlockInfo) AppHash() []byte {
	return merkle.SimpleHashFromTwoHashes(lb.EvmStateHash, lb.ShareStateHash)
}

type App struct {
	agtypes.BaseApplication

	datadir string
	logger  *zap.Logger
	Config  *viper.Viper

	AngineHooks agtypes.Hooks
	AngineRef   *angine.Angine

	evmStateMtx     sync.RWMutex
	evmStateDb      ethdb.Database
	evmState        *ethstate.StateDB
	currentEvmState *ethstate.StateDB

	ShareStateMtx     sync.RWMutex
	ShareStateDB      db.DB
	ShareState        *ShareState
	currentShareState *ShareState

	currentHeader *ethtypes.Header
	chainConfig   *ethparams.ChainConfig
	blockChain    *ethcore.BlockChain

	receipts ethtypes.Receipts
}

var (
	EmptyTrieRoot = ethcmn.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	EthSigner     = ethtypes.HomesteadSigner{}
	IsHomestead   = true

	evmConfig = ethvm.Config{DisableGasMetering: false, EnableJit: true, ForceJit: true}
	big0      = big.NewInt(0)

	errQuitExecute = fmt.Errorf("quit executing block")
)

func makeCurrentHeader(block *agtypes.BlockCache, conf *viper.Viper) *ethtypes.Header {
	return &ethtypes.Header{
		ParentHash: ethcmn.HexToHash("0x00"),
		Difficulty: big0,
		GasLimit:   big.NewInt(conf.GetInt64("block_gaslimit")),
		Number:     ethparams.MainNetSpuriousDragon,
		Time:       big.NewInt(block.Header.Time),
		Coinbase:   ethcmn.BytesToAddress(block.Header.CoinBase),
	}
}

func OpenDatabase(datadir string, name string, cache int, handles int) (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(filepath.Join(datadir, name), cache, handles)
}

func NewApp(logger *zap.Logger, config *viper.Viper) (*App, error) {
	app := &App{
		datadir:     config.GetString("db_dir"),
		chainConfig: new(ethparams.ChainConfig),
		logger:      logger,

		Config: config,
	}

	app.AngineHooks = agtypes.Hooks{
		OnNewRound: agtypes.NewHook(app.OnNewRound),
		OnCommit:   agtypes.NewHook(app.OnCommit),
		OnPrevote:  agtypes.NewHook(app.OnPrevote),
		OnExecute:  agtypes.NewHook(app.OnExecute),
	}

	appconfig.LoadDefaultAngineConfig(config)

	var err error
	if err = app.BaseApplication.InitBaseApplication(APP_NAME, app.datadir); err != nil {
		app.logger.Error("InitBaseApplication error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	if app.evmStateDb, err = OpenDatabase(app.datadir, "chaindata", DatabaseCache, DatabaseHandles); err != nil {
		app.logger.Error("OpenDatabase error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}
	if app.ShareStateDB, err = db.NewGoLevelDB("powerstate", app.datadir); err != nil {
		cmn.PanicCrisis(err)
	}

	return app, nil
}

func (app *App) Start() (err error) {
	lastBlock := &LastBlockInfo{
		Height:         0,
		EvmStateHash:   make([]byte, 0),
		ShareStateHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if err != nil {
		app.logger.Error("fail to load last block", zap.Error(err))
		return
	}

	// Load evm state when starting
	trieRoot := EmptyTrieRoot
	if len(lastBlock.EvmStateHash) > 0 {
		trieRoot = ethcmn.BytesToHash(lastBlock.EvmStateHash)
	}
	if app.evmState, err = ethstate.New(trieRoot, app.evmStateDb); err != nil {
		app.Stop()
		app.logger.Error("fail to new ethstate", zap.Error(err))
		return
	}

	// Load power state when starting
	app.ShareState = NewShareState(app.ShareStateDB)
	app.ShareState.Load(lastBlock.ShareStateHash)

	return nil
}

func (app *App) Stop() {
	app.BaseApplication.Stop()
	app.evmStateDb.Close()
	app.ShareStateDB.Close()
}

func (app *App) GetAngineHooks() agtypes.Hooks {
	return app.AngineHooks
}

func (app *App) CompatibleWithAngine() {}

func (app *App) OnNewRound(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	return agtypes.NewRoundResult{}, nil
}

func (app *App) OnPrevote(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	return nil, nil
}

func (app *App) OnExecute(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	app.currentEvmState = app.evmState.DeepCopy()
	app.currentShareState = app.ShareState.Copy()

	// genesis block
	if height == 1 {
		_, vSet := app.AngineRef.GetValidators()
		app.RegisterValidators(vSet)
	}

	currentHeader := makeCurrentHeader(block, app.Config)
	blockHash := ethcmn.BytesToHash(block.Hash())

	var err error
	var res agtypes.ExecuteResult
	exeWithCPUSerialVeirfy(block.Data.Txs, nil,
		func(index int, raw []byte, tx *types.BlockTx) {
			switch {
			case bytes.HasPrefix(raw, types.TxTagAppEvm):
				_, _, err = app.ExecuteEVMTx(currentHeader, blockHash, tx, index)
			case bytes.HasPrefix(raw, types.TxTagAngineInit):
				_, _, err = app.ExecuteAppInitTx(block, raw, index)
			case bytes.HasPrefix(raw, types.TxTagAppEco):
				_, _, err = app.ExecuteAppEcoTx(block, tx, index)
			default:
				return
			}

			if err != nil {
				res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: raw, Error: err})
			} else {
				res.ValidTxs = append(res.ValidTxs, raw)
			}
		}, func(bs []byte, err error) {
			res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: bs, Error: err})
		})

	// block rewards
	err = app.doCoinbaseTx(block)
	if err != nil {
		return nil, err
	}

	return res, err
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateMtx is still needed
func (app *App) OnCommit(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	evmStateHash, err := app.currentEvmState.Commit(StateRemoveEmptyObj)
	if err != nil {
		app.logger.Error("fail to commit evmState", zap.Error(err))
		return nil, err
	}

	ShareStateHash, err := app.currentShareState.Commit()
	if err != nil {
		app.logger.Error("fail to commit ShareState", zap.Error(err))
		return nil, err
	}

	lb := LastBlockInfo{Height: height, EvmStateHash: evmStateHash.Bytes(), ShareStateHash: ShareStateHash}
	app.SaveLastBlock(lb)
	rHash := app.SaveReceipts()

	app.SaveStateRootHashs(evmStateHash.Bytes(), ShareStateHash)

	// Reload status
	app.evmStateMtx.Lock()
	app.evmState, err = app.evmState.New(evmStateHash)
	app.evmStateMtx.Unlock()
	if err != nil {
		app.logger.Error("fail to new evmState", zap.Error(err))
		return nil, err
	}

	app.ShareStateMtx.Lock()
	app.ShareState.Reload(ShareStateHash)
	app.ShareStateMtx.Unlock()

	app.receipts = app.receipts[:0]

	fmt.Println("height:", height, "power size:", app.ShareState.Size())
	fmt.Printf("appHash:%X\n", lb.AppHash())

	return agtypes.CommitResult{
		AppHash:      lb.AppHash(),
		ReceiptsHash: rHash,
	}, nil
}

func (app *App) SaveReceipts() []byte {
	savedReceipts := make([][]byte, 0, len(app.receipts))
	receiptBatch := app.evmStateDb.NewBatch()

	for _, receipt := range app.receipts {
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

func (app *App) SaveStateRootHashs(evmStateHash, ShareStateHash []byte) {
	type stateRoot struct {
		EvmStateHash   []byte
		ShareStateHash []byte
	}
	sr := stateRoot{
		EvmStateHash:   evmStateHash,
		ShareStateHash: ShareStateHash,
	}
	srb, err := msgpack.Marshal(sr)
	if err != nil {
		cmn.PanicCrisis(err)
	}
	app.Database.SetSync([]byte(fmt.Sprintf("%s%d", stateRootPrefix, app.AngineRef.Height())), srb)
}

func (app *App) Info() (resInfo agtypes.ResultInfo) {
	lb := &LastBlockInfo{}
	if res, err := app.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}
	resInfo.LastBlockAppHash = lb.AppHash()
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "alpha 0.2"
	resInfo.Data = "evm-1.5.9"
	return
}
