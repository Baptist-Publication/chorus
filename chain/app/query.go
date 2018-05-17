package app

import (
	"bytes"
	"math/big"
	"time"

	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	ethcmn "github.com/Baptist-Publication/chorus/eth/common"
	ethcore "github.com/Baptist-Publication/chorus/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	ethvm "github.com/Baptist-Publication/chorus/eth/core/vm"
	ethparams "github.com/Baptist-Publication/chorus/eth/params"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	"github.com/Baptist-Publication/chorus/types"
)

func (app *App) Query(query []byte) agtypes.Result {
	if len(query) == 0 {
		return agtypes.NewResultOK([]byte{}, "Empty query")
	}
	var res agtypes.Result
	action := query[0]
	load := query[1:]
	switch action {
	case types.QueryTypeContract:
		res = app.queryContract(load)
	case types.QueryTypeNonce:
		res = app.queryNonce(load)
	case types.QueryTypeBalance:
		res = app.queryBalance(load)
	case types.QueryTypeReceipt:
		res = app.queryReceipt(load)
	case types.QueryTypeContractExistance:
		res = app.queryContractExistence(load)
	case types.QueryTypeShare:
		res = app.queryShare(load)
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

	app.evmStateMtx.RLock()
	hashBytes := app.evmState.GetCodeHash(*contractAddr).Bytes()
	app.evmStateMtx.RUnlock()

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
		GasLimit:   big.NewInt(app.Config.GetInt64("block_gaslimit")),
		Number:     ethparams.MainNetSpuriousDragon,
		Time:       big.NewInt(time.Now().Unix()),
	}
	txMsg, _ := tx.AsMessage()
	envCxt := ethcore.NewEVMContext(txMsg, fakeHeader, nil)

	app.evmStateMtx.Lock()
	defer app.evmStateMtx.Unlock()
	vmEnv := ethvm.NewEVM(envCxt, app.evmState.Copy(), app.chainConfig, evmConfig)
	gpl := new(ethcore.GasPool).AddGas(ethcmn.MaxBig)
	res, _, err := ethcore.ApplyMessage(vmEnv, txMsg, gpl) // we don't care about gasUsed
	if err != nil {
		// logger.Debug("transition error", err)
	}

	return agtypes.NewResultOK(res, "")
}

func (app *App) queryNonce(addrBytes []byte) agtypes.Result {
	if len(addrBytes) != 20 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.evmStateMtx.RLock()
	nonce := app.evmState.GetNonce(addr)
	app.evmStateMtx.RUnlock()

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

	app.evmStateMtx.RLock()
	balance := app.evmState.GetBalance(addr)
	app.evmStateMtx.RUnlock()

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
		return agtypes.NewError(pbtypes.CodeType_InternalError, "Failed to get receipt for tx:"+string(key))
	}
	return agtypes.NewResultOK(data, "")
}

func (app *App) queryShare(pubkeyBytes []byte) agtypes.Result {
	if len(pubkeyBytes) == 0 {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, "Invalid address")
	}

	app.ShareStateMtx.RLock()
	share := app.ShareState.GetShareAccount(pubkeyBytes)
	app.ShareStateMtx.RUnlock()

	res := types.QueryShareResult{}
	if share == nil {
		res.ShareBalance = big0
		res.ShareGuaranty = big0
		res.GHeight = 0
	} else {
		res.ShareBalance = share.ShareBalance
		res.ShareGuaranty = share.ShareGuaranty
		res.GHeight = uint64(share.GHeight)
	}

	data, err := rlp.EncodeToBytes(res)
	if err != nil {
		return agtypes.NewError(pbtypes.CodeType_InternalError, err.Error())
	}

	return agtypes.NewResultOK(data, "")
}

func (app *App) CheckTx(bs []byte) error {
	txBytes := agtypes.UnwrapTx(bs)

	switch {
	case bytes.HasPrefix(bs, types.TxTagAppEvm):
		return app.CheckEVMTx(txBytes)
	case bytes.HasPrefix(bs, types.TxTagAppEco):
		return app.CheckEcoTx(txBytes)
	}

	return nil
}
