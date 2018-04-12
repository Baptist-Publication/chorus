package app

import (
	agtypes "github.com/Baptist-Publication/angine/types"
	ethcmn "github.com/Baptist-Publication/chorus/src/eth/common"
	ethcore "github.com/Baptist-Publication/chorus/src/eth/core"
	ethtypes "github.com/Baptist-Publication/chorus/src/eth/core/types"
	"github.com/Baptist-Publication/chorus/src/eth/rlp"
)

var (
	ReceiptsPrefix = []byte("receipts-")

	EVMTag   = []byte{'e', 'v', 'm'}
	EVMTxTag = append(EVMTag, 0x01)
)

// ExecuteEVMTx execute tx one by one in the loop, without lock, so should always be called between Lock() and Unlock() on the *stateDup
func (app *App) ExecuteEVMTx(header *ethtypes.Header, blockHash ethcmn.Hash, bs []byte, txIndex int) (hash []byte, err error) {
	state := app.currentEvmState
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
		app.receipts = append(app.receipts, receipt)
	}

	return tx.Hash().Bytes(), err
}
