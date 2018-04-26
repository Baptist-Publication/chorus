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
	"errors"
	"reflect"
	"time"

	blkpb "github.com/Baptist-Publication/chorus/angine/protos/blockchain"
	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p"
	"github.com/Baptist-Publication/chorus/module/xlib/def"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	BlockchainChannel      = byte(0x40)
	defaultChannelCapacity = 100
	defaultSleepIntervalMS = 500
	trySyncIntervalMS      = 100

	// stop syncing when last block's time is within this much of the system time.
	stopSyncingDurationMinutes = 10

	// ask for best height every 10s
	statusUpdateIntervalSeconds = 10

	// check if we should switch to consensus reactor
	switchToConsensusIntervalSeconds = 1

	maxBlockchainResponseSize = agtypes.MaxBlockSize + 2
)

var ErrNotFound = errors.New("leveldb not found")

type BlockVerifierFunc func(pbtypes.BlockID, def.INT, *agtypes.CommitCache) error
type BlockExecuterFunc func(*agtypes.BlockCache, *agtypes.PartSet, *agtypes.CommitCache) error
type ValidatorSetorFunc func(*agtypes.BlockCache)

// BlockchainReactor handles long-term catchup syncing.
type BlockchainReactor struct {
	p2p.BaseReactor

	config     *viper.Viper
	store      *BlockStore
	pool       *BlockPool
	fastSync   bool
	requestsCh chan BlockRequest
	timeoutsCh chan string
	lastBlock  *pbtypes.Block

	blockVerifier  BlockVerifierFunc
	blockExecuter  BlockExecuterFunc
	validatorSetor ValidatorSetorFunc

	evsw agtypes.EventSwitch

	logger *zap.Logger
}

func NewBlockchainReactor(logger *zap.Logger, config *viper.Viper, lastBlockHeight def.INT, store *BlockStore, fastSync bool) *BlockchainReactor {

	if lastBlockHeight == store.Height()-1 {
		store.height -= 1 // XXX HACK, make this better
	}
	if lastBlockHeight != store.Height() {
		PanicSanity(Fmt("state (%v) and store (%v) height mismatch", lastBlockHeight, store.Height()))
	}
	requestsCh := make(chan BlockRequest, defaultChannelCapacity)
	timeoutsCh := make(chan string, defaultChannelCapacity)
	pool := NewBlockPool(
		logger,
		store.Height()+1,
		requestsCh,
		timeoutsCh,
	)
	bcR := &BlockchainReactor{
		config:     config,
		store:      store,
		pool:       pool,
		fastSync:   fastSync,
		requestsCh: requestsCh,
		timeoutsCh: timeoutsCh,
		logger:     logger,
	}
	bcR.BaseReactor = *p2p.NewBaseReactor(logger, "BlockchainReactor", bcR)
	return bcR
}

func (bcR *BlockchainReactor) SetBlockVerifier(v BlockVerifierFunc) {
	bcR.blockVerifier = v
}

func (bcR *BlockchainReactor) SetBlockExecuter(x BlockExecuterFunc) {
	bcR.blockExecuter = x
}

func (bcR *BlockchainReactor) SetStateValidator(f ValidatorSetorFunc) {
	bcR.validatorSetor = f
}

func (bcR *BlockchainReactor) OnStart() error {
	bcR.BaseReactor.OnStart()

	if bcR.fastSync {
		_, err := bcR.pool.Start()
		if err != nil {
			return err
		}
		go bcR.poolRoutine()
	}
	return nil
}

func (bcR *BlockchainReactor) OnStop() {
	bcR.BaseReactor.OnStop()
	bcR.pool.Stop()
}

// Implements Reactor
func (bcR *BlockchainReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID:                BlockchainChannel,
			Priority:          5,
			SendQueueCapacity: 100,
		},
	}
}

// Implements Reactor
func (bcR *BlockchainReactor) AddPeer(peer *p2p.Peer) {
	// Send peer our state.
	peer.SendBytes(BlockchainChannel,
		blkpb.MarshalDataToBlkMsg(&blkpb.StatusResponseMessage{
			Height: bcR.store.Height(),
		}))
}

// Implements Reactor
func (bcR *BlockchainReactor) RemovePeer(peer *p2p.Peer, reason interface{}) {
	// Remove peer from the pool.
	bcR.pool.RemovePeer(peer.Key)
}

// Implements Reactor
func (bcR *BlockchainReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	msg, err := blkpb.UnmarshalBlkMsg(msgBytes)
	if err != nil {
		bcR.logger.Warn("Error decoding message", zap.String("error", err.Error()))
		return
	}

	bcR.logger.Sugar().Debugw("Receive", "src", src, "chID", chID, "msg", msg)

	switch msg := msg.(type) {
	case *blkpb.BlockRequestMessage:
		// Got a request for a block. Respond with block if we have it.
		block := bcR.store.LoadBlock(msg.Height)
		if block != nil {
			queued := src.TrySendBytes(BlockchainChannel,
				blkpb.MarshalDataToBlkMsg(&blkpb.BlockResponseMessage{
					Block: block.Block,
				}))
			if !queued {
				// queue is full, just ignore.
			}
		} else {
			// TODO peer is asking for things we don't have.
		}
	case *blkpb.BlockResponseMessage:
		// Got a block.
		bcR.validatorSetor(agtypes.MakeBlockCache(msg.Block))
		bcR.pool.AddBlock(src.Key, msg.Block, len(msgBytes))
	case *blkpb.StatusRequestMessage:
		// Send peer our state.
		queued := src.TrySendBytes(BlockchainChannel,
			blkpb.MarshalDataToBlkMsg(&blkpb.StatusResponseMessage{
				Height: bcR.store.Height(),
			}))
		if !queued {
			// sorry
		}
	case *blkpb.StatusResponseMessage:
		// Got a peer status. Unverified.
		bcR.pool.SetPeerHeight(src.Key, msg.Height)
	case *blkpb.BlockHeaderRequestMessage:
		meta := bcR.store.LoadBlockMeta(msg.Height)
		if meta != nil && meta.Header != nil {
			queued := src.TrySendBytes(BlockchainChannel,
				blkpb.MarshalDataToBlkMsg(&blkpb.BlockHeaderResponseMessage{
					Header: meta.Header,
				}))
			if !queued {
				// queue is full, just ignore.
			}
		}
	default:
		bcR.logger.Warn(Fmt("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
// (Except for the SYNC_LOOP, which is the primary purpose and must be synchronous.)
func (bcR *BlockchainReactor) poolRoutine() {
	trySyncTicker := time.NewTicker(trySyncIntervalMS * time.Millisecond)
	statusUpdateTicker := time.NewTicker(statusUpdateIntervalSeconds * time.Second)
	switchToConsensusTicker := time.NewTicker(switchToConsensusIntervalSeconds * time.Second)

FOR_LOOP:
	for {
		select {
		case request := <-bcR.requestsCh: // chan BlockRequest
			peer := bcR.Switch.Peers().Get(request.PeerID)
			if peer == nil {
				continue FOR_LOOP // Peer has since been disconnected.
			}
			queued := peer.TrySendBytes(BlockchainChannel,
				blkpb.MarshalDataToBlkMsg(&blkpb.BlockRequestMessage{
					Height: request.Height,
				}))
			if !queued {
				// We couldn't make the request, send-queue full.
				// The pool handles timeouts, just let it go.
				continue FOR_LOOP
			}
		case peerID := <-bcR.timeoutsCh:
			// Peer timed out.
			peer := bcR.Switch.Peers().Get(peerID)
			if peer != nil {
				bcR.Switch.StopPeerForError(peer, errors.New("BlockchainReactor Timeout"))
			}
		case _ = <-statusUpdateTicker.C:
			// ask for status updates
			go bcR.BroadcastStatusRequest()
		case _ = <-switchToConsensusTicker.C:
			height, numPending, _ := bcR.pool.GetStatus()
			outbound, inbound, _ := bcR.Switch.NumPeers()
			bcR.logger.Debug("Consensus ticker", zap.Int32("numPending", numPending), zap.Int("total", len(bcR.pool.requesters)),
				zap.Int("outbound", outbound), zap.Int("inbound", inbound))
			if bcR.pool.IsCaughtUp() {
				bcR.logger.Info("Time to switch to consensus reactor!", zap.Int64("height", height))
				bcR.pool.Stop()
				agtypes.FireEventSwitchToConsensus(bcR.evsw)
				break FOR_LOOP
			}
		case _ = <-trySyncTicker.C: // chan time
			// This loop can be slow as long as it's doing syncing work.
		SYNC_LOOP:
			for i := 0; i < 10; i++ {
				// See if there are any blocks to sync.
				first, second := bcR.pool.PeekTwoBlocks()
				if first == nil || second == nil {
					// We need both to sync the first block.
					break SYNC_LOOP
				}
				firstParts := first.MakePartSet(bcR.config.GetInt64("block_part_size")) // TODO: put part size in parts header?
				firstPartsHeader := firstParts.Header()
				// Finally, verify the first block using the second's commit
				// NOTE: we can probably make this more efficient, but note that calling
				// first.Hash() doesn't verify the tx contents, so MakePartSet() is
				// currently necessary.
				if err := bcR.blockVerifier(pbtypes.BlockID{Hash: first.Hash(), PartsHeader: firstPartsHeader}, first.Header.Height, second.CommitCache()); err != nil {
					bcR.logger.Error("error in validation", zap.String("error", err.Error()))
					bcR.pool.RedoRequest(first.Header.Height)
					break SYNC_LOOP
				} else {
					bcR.pool.PopRequest()
					if err := bcR.blockExecuter(first, firstParts, second.CommitCache()); err != nil {
						// TODO This is bad, are we zombie?
						PanicQ(Fmt("Failed to process committed block (%d:%X): %v", first.Header.Height, first.Hash(), err))
					}
				}
			}
			continue FOR_LOOP
		case <-bcR.Quit:
			break FOR_LOOP
		}
	}
}

func (bcR *BlockchainReactor) BroadcastStatusResponse() error {
	bcR.Switch.BroadcastBytes(BlockchainChannel,
		blkpb.MarshalDataToBlkMsg(&blkpb.StatusResponseMessage{
			Height: bcR.store.Height(),
		}))
	return nil
}

func (bcR *BlockchainReactor) BroadcastStatusRequest() error {
	bcR.Switch.BroadcastBytes(BlockchainChannel,
		blkpb.MarshalDataToBlkMsg(&blkpb.StatusRequestMessage{
			Height: bcR.store.Height(),
		}))
	return nil
}

// implements events.Eventable
func (bcR *BlockchainReactor) SetEventSwitch(evsw agtypes.EventSwitch) {
	bcR.evsw = evsw
}

//-----------------------------------------------------------------------------
