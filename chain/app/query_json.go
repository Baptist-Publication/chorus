package app

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"bytes"

	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	ethcmn "github.com/Baptist-Publication/chorus/eth/common"
	"github.com/Baptist-Publication/chorus/eth/common/number"
	ethcore "github.com/Baptist-Publication/chorus/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	ethvm "github.com/Baptist-Publication/chorus/eth/core/vm"
	ethparams "github.com/Baptist-Publication/chorus/eth/params"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
	"github.com/pkg/errors"
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

	return agtypes.ResultQueryBalance{Code: pbtypes.CodeType_OK, Balance: (*number.BigNumber)(balance)}
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
		res.ShareBalance = (*number.BigNumber)(big0)
		res.ShareGuaranty = (*number.BigNumber)(big0)
		res.GHeight = (*number.BigNumber)(big0)
	} else {
		res.ShareBalance = (*number.BigNumber)(share.ShareBalance)
		res.ShareGuaranty = (*number.BigNumber)(share.ShareGuaranty)
		res.GHeight = (*number.BigNumber)(big.NewInt(0).SetUint64(share.GHeight))
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
		return agtypes.ResultQueryReceipt{Code: pbtypes.CodeType_InternalError, Log: "Fail to get tx: " + fmt.Sprintf("%X", txHashBytes)}
	}

	receipt := &ethtypes.ReceiptForStorage{}
	err = rlp.DecodeBytes(data, receipt)
	if err != nil {
		return agtypes.ResultQueryReceipt{Code: pbtypes.CodeType_InternalError, Log: "Fail to decode receipt for tx: " + fmt.Sprintf("%X", txHashBytes)}
	}

	return agtypes.ResultQueryReceipt{Code: pbtypes.CodeType_OK,
		Receipt: (*ethtypes.Receipt)(receipt)}
}

func GetTx(blockData *pbtypes.Data, hash []byte) (agtypes.Tx, error) {
	txs := agtypes.BytesToTxSlc(blockData.Txs)
	for _, tx := range txs {
		if bytes.Equal(tx.TxHash(), hash) {
			// found
			return tx, nil
		}
	}
	return agtypes.Tx{}, errors.New("Transaction not found")
}

// DecodeTx converts a raw []byte tx to a struct.
// Payload inside the ResultBlockTx may vary.
func DecodeTx(tx agtypes.Tx) (*agtypes.ResultBlockTx, error) {
	blockTxBytes := agtypes.UnwrapTx(tx)
	blockTx := types.BlockTx{}
	err := tools.FromBytes(blockTxBytes, &blockTx)
	if err != nil {
		return nil, errors.New("Failed to decode tx")
	}
	result := &agtypes.ResultBlockTx{
		GasLimit:  (*number.BigNumber)(blockTx.GasLimit),
		GasPrice:  (*number.BigNumber)(blockTx.GasPrice),
		Sender:    blockTx.Sender,
		Nonce:     blockTx.Nonce,
		Signature: blockTx.Signature,
	}
	switch {
	case bytes.HasPrefix(tx, types.TxTagAppEvmCommon):
		result.Tx_Type = "TxEvmCommon"
		payload := new(types.TxEvmCommon)
		tools.FromBytes(blockTx.Payload, payload)
		result.Payload = new(agtypes.ResultTxEvmCommon).Adapt(payload)
	case bytes.HasPrefix(tx, types.TxTagAppEcoShareTransfer):
		result.Tx_Type = "TxShareTransfer"
		payload := new(types.TxShareTransfer)
		tools.FromBytes(blockTx.Payload, payload)
		result.Payload = new(agtypes.ResultTxShareTransfer).Adapt(payload)
	case bytes.HasPrefix(tx, types.TxTagAppEcoGuarantee):
		result.Tx_Type = "TxShareGuarantee"
		payload := new(types.TxShareEco)
		tools.FromBytes(blockTx.Payload, payload)
		result.Payload = new(agtypes.ResultTxShareEco).Adapt(payload)
	case bytes.HasPrefix(tx, types.TxTagAppEcoRedeem):
		result.Tx_Type = "TxShareRedeem"
		payload := new(types.TxShareEco)
		tools.FromBytes(blockTx.Payload, payload)
		result.Payload = new(agtypes.ResultTxShareEco).Adapt(payload)

	default:
		return nil, errors.New("Tx type not supported yet: " + fmt.Sprintf("%X", tx[0:3]))
	}

	return result, nil
}

func (app *App) QueryTx(txHashBytes []byte) agtypes.ResultQueryTx {
	receiptResponse := app.QueryReceipt(txHashBytes)
	if receiptResponse.Code != pbtypes.CodeType_OK {
		return agtypes.ResultQueryTx{Code: receiptResponse.Code, Log: receiptResponse.Log}
	}
	receipt := receiptResponse.Receipt
	blockc, _, err := app.AngineRef.GetBlock(receipt.Height.Int64())
	if err != nil {
		return agtypes.ResultQueryTx{Code: pbtypes.CodeType_InternalError, Log: err.Error()}
	}
	tx, err := GetTx(blockc.Data, txHashBytes)
	if err != nil {
		return agtypes.ResultQueryTx{Code: pbtypes.CodeType_InvalidTx, Log: err.Error()}
	}

	decodedTx, err := DecodeTx(tx)
	if err != nil {
		return agtypes.ResultQueryTx{Code: pbtypes.CodeType_InternalError, Log: err.Error()}
	}
	return agtypes.ResultQueryTx{Code: pbtypes.CodeType_OK, Tx: decodedTx}
}

func (app *App) QueryContract(rawtx []byte) agtypes.ResultQueryContract {
	tx := &types.BlockTx{}
	err := rlp.DecodeBytes(rawtx, tx)
	if err != nil {
		return agtypes.ResultQueryContract{Code: pbtypes.CodeType_EncodingError}
	}
	txbody := &types.TxEvmCommon{}
	if err := tools.FromBytes(tx.Payload, txbody); err != nil {
		return agtypes.ResultQueryContract{Code: pbtypes.CodeType_EncodingError}
	}
	from := ethcmn.BytesToAddress(tx.Sender)
	to := ethcmn.BytesToAddress(txbody.To)
	evmtx := ethtypes.NewTransaction(tx.Nonce, from, to, txbody.Amount, tx.GasLimit, big.NewInt(0), txbody.Load)
	fakeHeader := &ethtypes.Header{
		ParentHash: ethcmn.HexToHash("0x00"),
		Difficulty: big0,
		GasLimit:   big.NewInt(app.Config.GetInt64("block_gaslimit")),
		Number:     ethparams.MainNetSpuriousDragon,
		Time:       big.NewInt(time.Now().Unix()),
	}
	txMsg, _ := evmtx.AsMessage()
	envCxt := ethcore.NewEVMContext(txMsg, fakeHeader, nil)

	app.evmStateMtx.Lock()
	defer app.evmStateMtx.Unlock()
	vmEnv := ethvm.NewEVM(envCxt, app.evmState.Copy(), app.chainConfig, evmConfig)
	gpl := new(ethcore.GasPool).AddGas(ethcmn.MaxBig)
	res, _, err := ethcore.ApplyMessage(vmEnv, txMsg, gpl) // we don't care about gasUsed
	if err != nil {
		return agtypes.ResultQueryContract{Code: pbtypes.CodeType_InternalError}
	}
	return agtypes.ResultQueryContract{Code: pbtypes.CodeType_OK, Data: hex.EncodeToString(res)}
}
