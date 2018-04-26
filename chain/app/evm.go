package app

import (
	"fmt"
	"math/big"

	ethcmn "github.com/Baptist-Publication/chorus/eth/common"
	ethcore "github.com/Baptist-Publication/chorus/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/eth/core/types"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	"github.com/Baptist-Publication/chorus/tools"
	"github.com/Baptist-Publication/chorus/types"
)

var (
	ReceiptsPrefix = []byte("receipts-")
)

// ExecuteEVMTx execute tx one by one in the loop, without lock, so should always be called between Lock() and Unlock() on the *stateDup
func (app *App) ExecuteEVMTx(header *ethtypes.Header, blockHash ethcmn.Hash, tx *types.BlockTx, txIndex int) (hash []byte, usedGas *big.Int, err error) {
	stateSnapshot := app.currentEvmState.Snapshot()

	txBody := &types.TxEvmCommon{}
	if err = tools.TxFromBytes(tx.Payload, txBody); err != nil {
		return
	}

	txHash := tx.Hash()
	from := ethcmn.BytesToAddress(tx.Sender)
	to := ethcmn.BytesToAddress(txBody.To)

	evmTx := ethtypes.NewTransaction(tx.Nonce, from, to, txBody.Amount, tx.GasLimit, tx.GasPrice, txBody.Load)
	gp := new(ethcore.GasPool).AddGas(header.GasLimit)
	fmt.Println("remaining gas of gaspool: ", gp)
	app.currentEvmState.StartRecord(ethcmn.BytesToHash(txHash), blockHash, txIndex)
	receipt, usedGas, err := ethcore.ApplyTransaction(
		app.chainConfig,
		nil,
		gp,
		app.currentEvmState,
		header,
		evmTx,
		txHash,
		big0,
		evmConfig)

	if err != nil {
		app.currentEvmState.RevertToSnapshot(stateSnapshot)
		return
	}

	if receipt != nil {
		app.receipts = append(app.receipts, receipt)
	}

	return txHash, usedGas, err
}

func (app *App) CheckEVMTx(bs []byte) error {
	tx := new(types.BlockTx)
	err := rlp.DecodeBytes(bs, tx)
	if err != nil {
		return err
	}
	valid, err := tx.VerifySignature()
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid signature")
	}
	from := ethcmn.BytesToAddress(tx.Sender)
	app.evmStateMtx.Lock()
	defer app.evmStateMtx.Unlock()
	if app.evmState.GetNonce(from) > tx.Nonce {
		return fmt.Errorf("nonce too low")
	}
	// if app.evmState.GetBalance(from).Cmp(tx.Cost()) < 0 {
	// 	return fmt.Errorf("not enough funds")
	// }
	return nil
}
