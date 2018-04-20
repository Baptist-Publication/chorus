// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"

	"github.com/Baptist-Publication/chorus/src/eth/common"
	"github.com/Baptist-Publication/chorus/src/eth/core/state"
	"github.com/Baptist-Publication/chorus/src/eth/core/types"
	"github.com/Baptist-Publication/chorus/src/eth/core/vm"
	"github.com/Baptist-Publication/chorus/src/eth/crypto"
	"github.com/Baptist-Publication/chorus/src/eth/logger"
	"github.com/Baptist-Publication/chorus/src/eth/logger/glog"
	"github.com/Baptist-Publication/chorus/src/eth/params"
)

var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

type BlockChain struct{}

func (bc BlockChain) GetHeader(common.Hash, uint64) *types.Header {
	return nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header,
	tx *types.Transaction, txHash []byte, usedGas *big.Int, cfg vm.Config) (*types.Receipt, *big.Int, error) {
	msg, err := tx.AsMessage()
	if err != nil {
		return nil, nil, err
	}
	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	_, gas, err := ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, nil, err
	}

	// Update the state with pending changes
	usedGas.Add(usedGas, gas)
	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	// Edit by Kyli
	// receipt := types.NewReceipt(statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes(), usedGas)
	receipt := types.NewReceipt([]byte{}, usedGas)
	// receipt.TxHash = tx.Hash()
	receipt.TxHash = common.BytesToHash(txHash)
	receipt.GasUsed = new(big.Int).Set(gas)
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}

	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(common.BytesToHash(txHash))
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	glog.V(logger.Debug).Infoln(receipt)

	return receipt, gas, err
}

// AccumulateRewards credits the coinbase of the given block with the
// mining reward. The total reward consists of the static block reward
// and rewards for included uncles. The coinbase of each uncle block is
// also rewarded.
func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Header) {
	reward := new(big.Int).Set(BlockReward)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, BlockReward)
		r.Div(r, big8)
		statedb.AddBalance(uncle.Coinbase, r)

		r.Div(BlockReward, big32)
		reward.Add(reward, r)
	}
	statedb.AddBalance(header.Coinbase, reward)
}
