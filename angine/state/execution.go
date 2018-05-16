// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	cmn "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus/module/xlib/def"
)

var (
	tpsc = NewTPSCalculator(10)
)

//--------------------------------------------------
// Execute the block

// Execute the block to mutate State.
// Validates block and then executes Data.Txs in the block.
func (s *State) ExecBlock(eventSwitch agtypes.EventSwitch, block *agtypes.BlockCache, blockPartsHeader *pbtypes.PartSetHeader, round def.INT) error {
	// Validate the block.
	if err := s.validateBlock(block); err != nil {
		return ErrInvalidBlock(err)
	}

	// compute bitarray of validators that signed
	signed := commitBitArrayFromBlock(block)
	_ = signed // TODO send on begin block

	// copy the valset
	valSet := block.VSetCache().Copy()
	bheader := block.Header

	s.blockExecutable.BeginBlock(block, eventSwitch, blockPartsHeader)
	s.execBlock(eventSwitch, block, round)
	s.blockExecutable.EndBlock(block, eventSwitch, blockPartsHeader)

	// All good!
	// Update validator accums and set state variables
	nextValSet := s.getNextValSet(valSet, bheader.Height, round)
	nextValSet.IncrementAccum(1)
	s.SetBlockAndValidators(bheader, blockPartsHeader, valSet, nextValSet)

	// save state with updated height/blockhash/validators
	// but stale apphash, in case we fail between Commit and Save
	s.SaveIntermediate()

	return nil
}

func (s *State) getNextValSet(currnet *agtypes.ValidatorSet, height, round def.INT) *agtypes.ValidatorSet {
	var valSet *agtypes.ValidatorSet

	// we re-elect every 500 blocks
	if s.valSetLoader != nil {
		valSet = s.valSetLoader(height, round)
		if valSet == nil || len(valSet.Validators) == 0 {
			return currnet.Copy()
		}

		// Print
		fmt.Println("Election happend in height ", height)
		count := 0
		for _, v := range valSet.Validators {
			pk := v.GetPubKey().(*crypto.PubKeyEd25519)
			fmt.Printf("[%X-%4d] ", pk[:2], v.VotingPower)
			count++
			if count%7 == 0 {
				fmt.Println()
			}
		}
		if count%7 != 0 {
			fmt.Println()
		}
		fmt.Println("=============================")
	} else {
		valSet = currnet.Copy()
	}

	return valSet
}

func (s *State) execBlock(eventSwitch agtypes.EventSwitch, block *agtypes.BlockCache, round def.INT) ([]*agtypes.ValidatorAttr, error) {
	// Run Txs of block
	bheader := block.Header
	ed := agtypes.NewEventDataHookExecute(bheader.Height, round, block)
	agtypes.FireEventHookExecute(eventSwitch, ed)
	res := <-ed.ResCh

	// flush events before handling errors
	eventCache := agtypes.NewEventCache(eventSwitch)
	s.blockExecutable.ExecBlock(block, eventCache, &res)

	for _, tx := range res.ValidTxs {
		txev := agtypes.EventDataTx{
			Tx:   tx,
			Code: pbtypes.CodeType_OK,
		}
		agtypes.FireEventTx(eventCache, txev)
	}

	for _, invalid := range res.InvalidTxs {
		txev := agtypes.EventDataTx{
			Tx:    invalid.Bytes,
			Code:  pbtypes.CodeType_InvalidTx,
			Error: invalid.Error.Error(),
		}
		agtypes.FireEventTx(eventCache, txev)
	}
	eventCache.Flush()

	if res.Error != nil {
		return nil, res.Error
	}

	tpsc.AddRecord(uint32(len(res.ValidTxs) + len(res.InvalidTxs)))
	tps := tpsc.TPS()

	if s.logger != nil {
		s.logger.Info("Executed block",
			zap.Int64("height", bheader.Height),
			zap.Int64("txs", bheader.NumTxs),
			zap.Int("valid", len(res.ValidTxs)),
			zap.Int("invalid", len(res.InvalidTxs)),
			zap.Int("extended", len(block.Data.ExTxs)),
			zap.Int("tps", tps))
	}

	return nil, nil
}

// return a bit array of validators that signed the last commit
// NOTE: assumes commits have already been authenticated
func commitBitArrayFromBlock(block *agtypes.BlockCache) *cmn.BitArray {
	signed := cmn.NewBitArray(len(block.LastCommit.Precommits))
	for i, precommit := range block.LastCommit.Precommits {
		if precommit != nil {
			signed.SetIndex(i, true) // val_.LastCommitHeight = block.Height - 1
		}
	}
	return signed
}

//-----------------------------------------------------
// Validate block

func (s *State) ValidateBlock(block *agtypes.BlockCache) error {
	return s.validateBlock(block)
}

func (s *State) validateBlock(block *agtypes.BlockCache) error {
	// Basic block validation.
	err := block.ValidateBasic(s.ChainID, s.LastBlockHeight, s.LastBlockID, s.LastBlockTime, s.AppHash, s.ReceiptsHash, s.LastNonEmptyHeight)
	if err != nil {
		return err
	}

	bheader := block.GetHeader()
	precmmt := block.CommitCache().GetPrecommits()
	// Validate block LastCommit.
	if bheader.Height == 1 {
		if len(precmmt) != 0 {
			return errors.New("Block at height 1 (first block) should have no LastCommit precommits")
		}
	} else {
		if len(precmmt) != s.LastValidators.Size() {
			return errors.New(cmn.Fmt("Invalid block commit size. Expected %v, got %v",
				s.LastValidators.Size(), len(precmmt)))
		}
		err := s.LastValidators.VerifyCommit(
			s.ChainID, s.LastBlockID, bheader.Height-1, block.CommitCache())
		if err != nil {
			return err
		}
	}

	return nil
}

// ApplyBlock executes the block, then commits and updates the mempool atomically
func (s *State) ApplyBlock(eventSwitch agtypes.EventSwitch, block *agtypes.BlockCache, partsHeader *pbtypes.PartSetHeader, mempool agtypes.IMempool, round def.INT) error {
	// Run the block on the State:
	// + update validator sets
	// + run txs on the proxyAppConn
	err := s.ExecBlock(eventSwitch, block, partsHeader, round)
	if err != nil {
		return errors.New(cmn.Fmt("Exec failed for application: %v", err))
	}
	// lock mempool, commit state, update mempoool
	err = s.CommitStateUpdateMempool(eventSwitch, block, mempool, round)
	if err != nil {
		return errors.New(cmn.Fmt("Commit failed for application: %v", err))
	}

	return nil
}

// mempool must be locked during commit and update
// because state is typically reset on Commit and old txs must be replayed
// against committed state before new txs are run in the mempool, lest they be invalid
func (s *State) CommitStateUpdateMempool(eventSwitch agtypes.EventSwitch, block *agtypes.BlockCache, mempool agtypes.IMempool, round def.INT) error {
	bheader := block.GetHeader()
	mempool.Update(bheader.Height, append(agtypes.BytesToTxSlc(block.GetData().Txs),
		agtypes.BytesToTxSlc(block.GetData().ExTxs)...))
	ed := agtypes.NewEventDataHookCommit(bheader.Height, round, block)
	agtypes.FireEventHookCommit(eventSwitch, ed)
	res := <-ed.ResCh
	s.AppHash = res.AppHash
	s.ReceiptsHash = res.ReceiptsHash
	return nil
}

//----------------------------------------------------------------
// Handshake with app to sync to latest state of core by replaying blocks

// TODO: Should we move blockchain/store.go to its own package?
type BlockStore interface {
	Height() def.INT
	LoadBlock(height def.INT) *agtypes.BlockCache
	LoadBlockMeta(height def.INT) *pbtypes.BlockMeta
}

type Handshaker struct {
	config *viper.Viper
	state  *State
	store  BlockStore

	nBlocks int // number of blocks applied to the state
}

func NewHandshaker(config *viper.Viper, state *State, store BlockStore) *Handshaker {
	return &Handshaker{config, state, store, 0}
}

// TODO: retry the handshake/replay if it fails ?
func (h *Handshaker) Handshake() error {
	// blockHeight := int(res.LastBlockHeight)
	// appHash := res.LastBlockAppHash

	// TODO: check version

	// replay blocks up to the latest in the blockstore
	// err = h.ReplayBlocks(appHash, blockHeight)
	// if err != nil {
	// 	return errors.New(Fmt("Error on replay: %v", err))
	// }

	// Save the state
	// h.state.Save()

	// TODO: (on restart) replay mempool

	return nil
}
