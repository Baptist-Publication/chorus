package p2p

import (
	"reflect"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Baptist-Publication/chorus/angine/consensus"
	csspb "github.com/Baptist-Publication/chorus/angine/protos/consensus"
	p2ppb "github.com/Baptist-Publication/chorus/angine/protos/p2p"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p"
)

const (
	OptimiseChannel      = byte(0x50)
	CheckRequestChannel  = byte(0x51)
	CheckResponseChannel = byte(0x52)
	// TODO more channel

)

type P2PReactor struct {
	p2p.BaseReactor

	config  *viper.Viper
	evsw    agtypes.EventSwitch
	logger  *zap.Logger
	slogger *zap.SugaredLogger
}

func NewP2PReactor(logger *zap.Logger, config *viper.Viper) *P2PReactor {
	pR := &P2PReactor{
		config:  config,
		logger:  logger,
		slogger: logger.Sugar(),
	}

	pR.BaseReactor = *p2p.NewBaseReactor(logger, "P2PReactor", pR)
	return pR
}

func (pR *P2PReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID:                 OptimiseChannel,
			Priority:           5, // once we gossip the whole block there's nothing left to send until next height or round
			SendQueueCapacity:  100,
			RecvBufferCapacity: 50 * 4096,
		},
		&p2p.ChannelDescriptor{
			ID:                 CheckRequestChannel,
			Priority:           5, // once we gossip the whole block there's nothing left to send until next height or round
			SendQueueCapacity:  100,
			RecvBufferCapacity: 50 * 4096,
		},
		&p2p.ChannelDescriptor{
			ID:                 CheckResponseChannel,
			Priority:           5, // once we gossip the whole block there's nothing left to send until next height or round
			SendQueueCapacity:  100,
			RecvBufferCapacity: 50 * 4096,
		},
	}
}

// Implement AddPeer
func (pR *P2PReactor) AddPeer(peer *p2p.Peer) {
	go pR.checkMsgRepeatedRoutine(peer)
}

func (pR *P2PReactor) RemovePeer(peer *p2p.Peer, reason interface{}) {
	// nothing to do yet
}

func (pR *P2PReactor) checkBlockPartExists(height, round int64, index int32) bool {
	cssReator := pR.Switch.Reactor("CONSENSUS").(*consensus.ConsensusReactor)
	conS := cssReator.GetConsensusState()
	if conS == nil {
		pR.logger.Warn(" ConsensusState is nil")
		return false
	}
	if conS.Height != height {
		return false
	}
	return conS.ProposalBlockParts.PartExists(int(index))
}

func (pR *P2PReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	// TODO check logic
	if !pR.IsRunning() {
		pR.slogger.Debugw("Receive", "src", src, "chId", chID, "bytes", msgBytes)
		return
	}
	msg, err := p2ppb.UnmarshalP2pMsg(msgBytes)
	if err != nil {
		pR.slogger.Warnw("Error decoding message", "src", src, "chId", chID, "msg", msg, "error", err, "bytes", msgBytes)
		// TODO punish peer?
		return
	}

	switch chID {
	case CheckRequestChannel:
		switch msg := msg.(type) {
		case *p2ppb.BlockPartMessage:
			var exists bool
			exists = pR.checkBlockPartExists(msg.Height, msg.Round, msg.Index)
			blockPartMsg := &p2ppb.BlockPartMessage{
				MsgId:  msg.MsgId,
				Height: msg.Height, // This tells peer that this part applies to us.
				Round:  msg.Round,  // This tells peer that this part applies to us.
				Index:  msg.Index,
			}
			if exists {
				blockPartMsg.Response = p2ppb.MsgResponseType_DataExists
			} else {
				blockPartMsg.Response = p2ppb.MsgResponseType_DataExists
			}
			msgBytes := p2ppb.MarshalDataToChkMsg(blockPartMsg)
			src.PureTrySendBytes(CheckResponseChannel, msgBytes)
		default:
			pR.slogger.Warnf("Unknown message type %T", reflect.TypeOf(msg))
		}

	case CheckResponseChannel:
		switch msg := msg.(type) {
		case *p2ppb.BlockPartMessage:
			msgId := msg.MsgId
			// TODO get msgID from msgBytes
			if cmCh, ok := src.CheckRespChSet[msgId]; ok {
				msgRepeated := false
				// TODO get check result from msgBytes
				if msg.GetResponse() == p2ppb.MsgResponseType_DataExists {
					msgRepeated = true
				} else {
					msgRepeated = false
				}
				cmCh <- msgRepeated
			}
		default:
			pR.slogger.Warnf("Unknown message type %T", reflect.TypeOf(msg))

		}

	default:
		pR.slogger.Warnf("Unknown chId %X", chID)

	}

	return
}

func (pR *P2PReactor) checkMsgRepeatedRoutine(peer *p2p.Peer) {
	for {
		select {
		case cr := <-peer.CheckMsgCh:
			switch cr.ChID {
			case consensus.DataChannel:
				msg, err := csspb.UnmarshalCssMsg(cr.Msg)
				if err != nil {
					pR.slogger.Warnw("Error decoding message", "src", peer.Key, "chId", cr.ChID, "error", err, "bytes", cr.Msg)
					// TODO punish peer?
					continue
				}
				switch msg := msg.(type) {
				case *csspb.BlockPartMessage:
					blockPartMsg := &p2ppb.BlockPartMessage{
						MsgId:  cr.MsgID,
						Height: msg.Height, // This tells peer that this part applies to us.
						Round:  msg.Round,  // This tells peer that this part applies to us.
						Index:  msg.Part.Index,
					}
					msgBytes := p2ppb.MarshalDataToChkMsg(blockPartMsg)
					peer.PureSendBytes(CheckRequestChannel, msgBytes)
					pR.slogger.Infow(" send checkmsg", "src", peer.Key, "chId", cr.ChID, "error", err, "bytes", cr.Msg)
				default:
					pR.DirectResponse(peer, cr.MsgID)
				}

				// TODO
			//case mempool.MempoolChannel:
			default:
				pR.DirectResponse(peer, cr.MsgID)
				// TODO
			}
		}
	}
}

func (pR *P2PReactor) DirectResponse(peer *p2p.Peer, msgId string) {
	if cmCh, ok := peer.CheckRespChSet[msgId]; ok {
		// TODO get check result from msgByte
		cmCh <- false
	}
}

// implements events.Eventable
func (pR *P2PReactor) SetEventSwitch(evsw agtypes.EventSwitch) {
	pR.evsw = evsw
}
