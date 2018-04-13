package app

import (
	"bytes"
	"math/big"
	"time"

	pbtypes "github.com/Baptist-Publication/angine/protos/types"
	agtypes "github.com/Baptist-Publication/angine/types"
	ethcmn "github.com/Baptist-Publication/chorus/src/eth/common"
	ethcore "github.com/Baptist-Publication/chorus/src/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/src/eth/core/types"
	ethvm "github.com/Baptist-Publication/chorus/src/eth/core/vm"
	ethparams "github.com/Baptist-Publication/chorus/src/eth/params"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
)

func (app *App) Query(query []byte) agtypes.Result {
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

func (app *App) queryContractExistence(load []byte) agtypes.Result {
	tx := new(ethtypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		// logger.Error("fail to decode tx: ", err)
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, err.Error())
	}
	contractAddr := tx.To()

	app.evmStateMtx.Lock()
	hashBytes := app.evmState.GetCodeHash(*contractAddr).Bytes()
	app.evmStateMtx.Unlock()

	if bytes.Equal(tx.Data(), hashBytes) {
		return agtypes.NewResultOK(append([]byte{}, byte(0x01)), "contract exists")
	}
	return agtypes.NewResultOK(append([]byte{}, byte(0x00)), "constract doesn't exist")
}

func (app *App) queryContract(load []byte) agtypes.Result {
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

	app.evmStateMtx.Lock()
	vmEnv := ethvm.NewEVM(envCxt, app.evmState.Copy(), app.chainConfig, evmConfig)
	gpl := new(ethcore.GasPool).AddGas(ethcmn.MaxBig)
	res, _, err := ethcore.ApplyMessage(vmEnv, txMsg, gpl) // we don't care about gasUsed
	if err != nil {
		// logger.Debug("transition error", err)
	}
	app.evmStateMtx.Unlock()

	return agtypes.NewResultOK(res, "")
}

func (app *App) queryNonce(addrBytes []byte) agtypes.Result {
	if len(addrBytes) != 20 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.evmStateMtx.Lock()
	nonce := app.evmState.GetNonce(addr)
	app.evmStateMtx.Unlock()

	data, err := rlp.EncodeToBytes(nonce)
	if err != nil {
		// logger.Warn("query error", err)
	}
	return agtypes.NewResultOK(data, "")
}

func (app *App) queryBalance(addrBytes []byte) agtypes.Result {
	if len(addrBytes) != 20 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.evmStateMtx.Lock()
	balance := app.evmState.GetBalance(addr)
	app.evmStateMtx.Unlock()

	data, err := rlp.EncodeToBytes(balance)
	if err != nil {
		// logger.Warn("query error", err)
	}
	return agtypes.NewResultOK(data, "")
}

func (app *App) queryReceipt(txHashBytes []byte) agtypes.Result {
	if len(txHashBytes) != 32 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid txhash")
	}
	key := append(ReceiptsPrefix, txHashBytes...)
	data, err := app.evmStateDb.Get(key)
	if err != nil {
		return agtypes.NewError(pbtypes.CodeType_InternalError, "fail to get receipt for tx:"+string(key))
	}
	return agtypes.NewResultOK(data, "")
}

func (app *App) CheckTx(bs []byte) error {
	txBytes := agtypes.UnwrapTx(bs)

	switch {
	case bytes.HasPrefix(bs, EVMTag):
		return app.CheckEVMTx(txBytes)
	case bytes.HasPrefix(bs, EcoTag):
		return app.CheckEcoTx(txBytes)
	}

	return nil
}
