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

package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	dbm "github.com/Baptist-Publication/chorus/module/lib/go-db"
	"github.com/Baptist-Publication/chorus/module/xlib/def"
)

/*
Simple low level store for blocks.

There are three types of information stored:
 - BlockMeta:   Meta information about each block
 - Block part:  Parts of each block, aggregated w/ PartSet
 - Commit:      The commit part of each block, for gossiping precommit votes

Currently the precommit signatures are duplicated in the Block parts as
well as the Commit.  In the future this may change, perhaps by moving
the Commit data outside the Block.

Panics indicate probable corruption in the data
*/
type BlockStore struct {
	db dbm.DB

	mtx    sync.RWMutex
	height def.INT
}

func NewBlockStore(db dbm.DB) *BlockStore {
	bsjson := LoadBlockStoreStateJSON(db)
	return &BlockStore{
		height: bsjson.Height,
		db:     db,
	}
}

// Height() returns the last known contiguous block height.
func (bs *BlockStore) Height() def.INT {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.height
}

func (bs *BlockStore) GetReader(key []byte) io.Reader {
	bytez := bs.db.Get(key)
	if bytez == nil {
		return nil
	}
	return bytes.NewReader(bytez)
}

func (bs *BlockStore) LoadBlock(height def.INT) *agtypes.BlockCache {
	var err error
	meta := bs.LoadBlockMeta(height)
	if meta == nil {
		return nil
	}
	bytez := []byte{}
	for i := 0; i < int(meta.PartsHeader.Total); i++ {
		part := bs.LoadBlockPart(height, i)
		bytez = append(bytez, part.Bytes...)
	}
	var block pbtypes.Block
	err = agtypes.UnmarshalData(bytez, &block)
	if err != nil {
		PanicCrisis(Fmt("Error reading block: %v", err))
	}
	return agtypes.MakeBlockCache(&block)
}

func (bs *BlockStore) LoadBlockPart(height def.INT, index int) *pbtypes.Part {
	var err error
	partBys := bs.db.Get(calcBlockPartKey(height, index))
	if len(partBys) == 0 {
		return nil
	}
	var part pbtypes.Part
	err = agtypes.UnmarshalData(partBys, &part)
	if err != nil {
		PanicCrisis(Fmt("Error reading block part: %v", err))
	}

	return &part
}

func (bs *BlockStore) LoadBlockMeta(height def.INT) *pbtypes.BlockMeta {
	return bs.loadBlockMeta(height, false)
}

func (bs *BlockStore) loadBlockMeta(height def.INT, archv bool) *pbtypes.BlockMeta {
	metaBys := bs.db.Get(calcBlockMetaKey(height))
	var meta pbtypes.BlockMeta
	err := agtypes.UnmarshalData(metaBys, &meta)
	if err != nil {
		PanicCrisis(Fmt("Error reading block meta: %v", err))
	}
	return &meta
}

// The +2/3 and other Precommit-votes for block at `height`.
// This Commit comes from block.LastCommit for `height+1`.
func (bs *BlockStore) LoadBlockCommit(height def.INT) *agtypes.CommitCache {
	var err error
	commitBys := bs.db.Get(calcBlockCommitKey(height))
	if len(commitBys) == 0 {
		return nil
	}
	var commit pbtypes.Commit
	err = agtypes.UnmarshalData(commitBys, &commit)
	if err != nil {
		PanicCrisis(Fmt("Error reading commit: %v", err))
	}
	return agtypes.NewCommitCache(&commit)
}

// NOTE: the Precommit-vote heights are for the block at `height`
func (bs *BlockStore) LoadSeenCommit(height def.INT) *agtypes.CommitCache {
	var err error
	commitBys := bs.db.Get(calcSeenCommitKey(height))
	if len(commitBys) == 0 {
		return nil
	}
	var commit pbtypes.Commit
	err = agtypes.UnmarshalData(commitBys, &commit)
	if err != nil {
		PanicCrisis(Fmt("Error reading commit: %v", err))
	}
	return agtypes.NewCommitCache(&commit)
}

// blockParts: Must be parts of the block
// seenCommit: The +2/3 precommits that were seen which committed at height.
//             If all the nodes restart after committing a block,
//             we need this to reload the precommits to catch-up nodes to the
//             most recent height.  Otherwise they'd stall at H-1.
func (bs *BlockStore) SaveBlock(block *agtypes.BlockCache, blockParts *agtypes.PartSet, seenCommit *agtypes.CommitCache) {
	height := block.GetHeader().GetHeight()
	if height != bs.Height()+1 {
		PanicSanity(Fmt("BlockStore can only save contiguous blocks. Wanted %v, got %v", bs.Height()+1, height))
	}
	if !blockParts.IsComplete() {
		PanicSanity(Fmt("BlockStore can only save complete block part sets"))
	}

	// Save block meta
	meta := agtypes.NewBlockMeta(block, blockParts)
	metaBytes, err := agtypes.MarshalData(meta)
	if err != nil {
		//TODO LOG
	}
	bs.db.Set(calcBlockMetaKey(height), metaBytes)
	// fmt.Println("====metaBytes:", len(metaBytes))

	// Save block parts
	for i := 0; i < int(blockParts.Total()); i++ {
		bs.saveBlockPart(height, i, blockParts.GetPart(i))
	}

	// Save block commit (duplicate and separate from the Block)
	blockCommitBytes, err := agtypes.MarshalData(block.Block.GetLastCommit())
	if err != nil {
		//TODO LOG
	}
	bs.db.Set(calcBlockCommitKey(height-1), blockCommitBytes)
	// fmt.Println("======commit bytes:", len(blockCommitBytes))

	// Save seen commit (seen +2/3 precommits for block)
	// NOTE: we can delete this at a later height
	seenCommitBytes, err := agtypes.MarshalData(seenCommit.Commit)
	if err != nil {
		//TODO LOG
	}
	bs.db.Set(calcSeenCommitKey(height), seenCommitBytes)
	// fmt.Println("======seen commit bytes:", len(seenCommitBytes))

	// Save new BlockStoreStateJSON descriptor
	BlockStoreStateJSON{Height: height}.Save(bs.db)

	// Done!
	bs.mtx.Lock()
	bs.height = height
	bs.mtx.Unlock()

	// Flush
	bs.db.SetSync(nil, nil)
}

func (bs *BlockStore) saveBlockPart(height def.INT, index int, part *pbtypes.Part) {
	if height != bs.Height()+1 {
		PanicSanity(Fmt("BlockStore can only save contiguous blocks. Wanted %v, got %v", bs.Height()+1, height))
	}
	partBytes, err := agtypes.MarshalData(part)
	if err != nil {
		// TODO err log
	}
	bs.db.Set(calcBlockPartKey(height, index), partBytes)
	// fmt.Println("====block part:", len(partBytes))
}

//-----------------------------------------------------------------------------

func calcBlockMetaKey(height def.INT) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

func calcBlockPartKey(height def.INT, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}

func calcBlockCommitKey(height def.INT) []byte {
	return []byte(fmt.Sprintf("C:%v", height))
}

func calcSeenCommitKey(height def.INT) []byte {
	return []byte(fmt.Sprintf("SC:%v", height))
}

//-----------------------------------------------------------------------------

var blockStoreKey = []byte("blockStore")

type BlockStoreStateJSON struct {
	Height       def.INT
	OriginHeight def.INT
}

func (bsj BlockStoreStateJSON) Save(db dbm.DB) {
	bsj.SaveByKey(blockStoreKey, db)
}

func (bsj BlockStoreStateJSON) SaveByKey(key []byte, db dbm.DB) {
	bytes, err := json.Marshal(bsj)
	if err != nil {
		PanicSanity(Fmt("Could not marshal state bytes: %v", err))
	}
	db.SetSync(key, bytes)
	// fmt.Println("=====block store state JSON", len(bytes))
}

func LoadBlockStoreStateJSON(db dbm.DB) BlockStoreStateJSON {
	bytes := db.Get(blockStoreKey)
	if bytes == nil {
		return BlockStoreStateJSON{
			Height:       0,
			OriginHeight: 0,
		}
	}
	bsj := BlockStoreStateJSON{}
	err := json.Unmarshal(bytes, &bsj)
	if err != nil {
		PanicCrisis(Fmt("Could not unmarshal bytes: %X", bytes))
	}
	return bsj
}

type NonEmptyBlockIterator struct {
	cursor     def.INT
	blockstore *BlockStore
}

func NewNonEmptyBlockIterator(store *BlockStore) *NonEmptyBlockIterator {
	height := store.Height()
	meta := store.LoadBlockMeta(height)
	var nonEmpty def.INT = 0
	if meta.Header.NumTxs == 0 {
		nonEmpty = meta.Header.LastNonEmptyHeight
	} else {
		nonEmpty = height
	}

	return &NonEmptyBlockIterator{
		cursor:     nonEmpty,
		blockstore: store,
	}
}

func (i *NonEmptyBlockIterator) Next() *agtypes.BlockCache {
	if i.cursor == 0 {
		return nil
	}
	block := i.blockstore.LoadBlock(i.cursor)
	i.cursor = block.Header.LastNonEmptyHeight
	return block
}

func (i *NonEmptyBlockIterator) HasMore() bool {
	return i.cursor != 0
}
