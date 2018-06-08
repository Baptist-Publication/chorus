package p2p

import (
	"go.uber.org/zap"
	"github.com/spf13/viper"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p"
	"github.com/Baptist-Publication/chorus/angine/consensus"
	"github.com/Baptist-Publication/chorus/angine/mempool"
)

const (
	OptimiseChannel = byte(0x50)
	// TODO more channel 

)

type P2PReactor struct {
	p2p.BaseReactor

	config  *viper.Viper
	evsw    agtypes.EventSwitch
	logger *zap.Logger
}

func NewP2PReactor(logger *zap.Logger, config *viper.Viper) *P2PReactor {
	pR := &P2PReactor{
		config: config,
		logger: logger,
	}
	pR.BaseReactor = *p2p.NewBaseReactor(logger, "P2PReactor", pR)
	return pR
}

func (pR *P2PReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID:		OptimiseChannel,
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

func (pR *P2PReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	// TODO check logic


	// TODO get msgID from msgBytes
	msgID := "blablabla"
	if cmCh, ok := src.CheckRespChSet[msgID]; ok {
		// TODO get check result from msgBytes
		msgRepeated := true
		cmCh <- msgRepeated
	} 
}

func (pR *P2PReactor) checkMsgRepeatedRoutine(peer *p2p.Peer) {
	for {
		select  {
		case cr := <- peer.CheckMsgCh:
			switch cr.ChID {
			case consensus.DataChannel:
				// TODO
			case mempool.MempoolChannel:
				// TODO
			} 


			// TODO implement a peer.Send
		}
	}
}

// implements events.Eventable
func (pR *P2PReactor) SetEventSwitch(evsw agtypes.EventSwitch) {
	pR.evsw = evsw
}









