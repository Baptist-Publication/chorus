package app

import (
	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	ethcmn "github.com/Baptist-Publication/chorus/eth/common"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	"fmt"
	"github.com/Baptist-Publication/chorus/eth/common/hexutil"
	"math/big"
)
// This file is for JSON interface
// Temporarily copied implementation from query.go but will not return an RLP encoded []byte
// It will return an interface{} directly.

func (app *App) QueryNonce(addrBytes []byte) agtypes.ResultQueryNonce {
	if len(addrBytes) != 20 {
		return agtypes.ResultQueryNonce{Code: pbtypes.CodeType_BaseInvalidInput, Log: "Invalid address"}
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.evmStateMtx.RLock()
	nonce := app.evmState.GetNonce(addr)
	app.evmStateMtx.RUnlock()

	return agtypes.ResultQueryNonce{Code: pbtypes.CodeType_OK, Nonce: nonce}
}

func (app *App) QueryBalance(addrBytes []byte) agtypes.ResultQueryBalance {
	if len(addrBytes) != 20 {
		return agtypes.ResultQueryBalance{Code: pbtypes.CodeType_BaseInvalidInput, Log: "Invalid address"}
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.evmStateMtx.RLock()
	balance := app.evmState.GetBalance(addr)
	app.evmStateMtx.RUnlock()

	return agtypes.ResultQueryBalance{Code: pbtypes.CodeType_OK, Balance: (*hexutil.Big)(balance)}
}

func (app *App) QueryShare(pubkeyBytes []byte) agtypes.ResultQueryShare {
	if len(pubkeyBytes) == 0 {
		return agtypes.ResultQueryShare{Code: pbtypes.CodeType_BaseInvalidInput, Log: "Invalid pubkey"}
	}

	app.ShareStateMtx.RLock()
	share := app.ShareState.GetShareAccount(pubkeyBytes)
	app.ShareStateMtx.RUnlock()

	res := agtypes.ResultQueryShare{Code: pbtypes.CodeType_OK}
	if share == nil {
		res.ShareBalance = (*hexutil.Big)(big0)
		res.ShareGuaranty = (*hexutil.Big)(big0)
		res.GHeight = (*hexutil.Big)(big0)
	} else {
		res.ShareBalance = (*hexutil.Big)(share.ShareBalance)
		res.ShareGuaranty = (*hexutil.Big)(share.ShareGuaranty)
		res.GHeight = (*hexutil.Big)(big.NewInt(0).SetUint64(share.GHeight))
	}

	return res
}

func (app *App) QueryReceipt(txHashBytes []byte) agtypes.ResultQueryReceipt {
	if len(txHashBytes) != 32 {
		return agtypes.ResultQueryReceipt{Code: pbtypes.CodeType_BaseInvalidInput, Log: "Invalid txhash"}
	}
	key := append(ReceiptsPrefix, txHashBytes...)
	data, err := app.evmStateDb.Get(key)
	if err != nil {
		return agtypes.ResultQueryReceipt{Code: pbtypes.CodeType_InternalError, Log: "Fail to get receipt for tx: " + fmt.Sprintf("%X", txHashBytes)}
	}

	receipt := &ethtypes.ReceiptForStorage{}
	err = rlp.DecodeBytes(data, receipt)
	if err != nil {
		return agtypes.ResultQueryReceipt{Code: pbtypes.CodeType_InternalError, Log: "Fail to decode receipt for tx: " + fmt.Sprintf("%X", txHashBytes)}
	}

	return agtypes.ResultQueryReceipt{Code: pbtypes.CodeType_OK,
		Receipt: (*ethtypes.Receipt)(receipt)}
}
