package app

import (
	"fmt"
	"math/big"

	ethcmn "github.com/Baptist-Publication/chorus/src/eth/common"
	ethcore "github.com/Baptist-Publication/chorus/src/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/src/eth/core/types"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
	"github.com/Baptist-Publication/chorus/src/tools"
	"github.com/Baptist-Publication/chorus/src/types"
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

	evmTx := ethtypes.NewTransaction(tx.Nonce, ethcmn.BytesToAddress(txBody.To), txBody.Amount, tx.GasLimit, tx.GasPrice, txBody.Load)
	gp := new(ethcore.GasPool).AddGas(header.GasLimit)
	fmt.Println("remaining gas of gaspool: ", gp)
	app.currentEvmState.StartRecord(evmTx.Hash(), blockHash, txIndex)
	receipt, usedGas, err := ethcore.ApplyTransaction(
		app.chainConfig,
		nil,
		gp,
		app.currentEvmState,
		header,
		evmTx,
		big0,
		evmConfig)

	if err != nil {
		app.currentEvmState.RevertToSnapshot(stateSnapshot)
		return
	}

	if receipt != nil {
		app.receipts = append(app.receipts, receipt)
	}

	return evmTx.Hash().Bytes(), usedGas, err
}

func (app *App) CheckEVMTx(bs []byte) error {
	tx := new(types.BlockTx)
	err := rlp.DecodeBytes(bs, tx)
	if err != nil {
		return err
	}
	valid, err := tools.TxVerifySignature(tx)
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
