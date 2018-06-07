package p2p

import (
	"go.uber.org/zap"
	"github.com/spf13/viper"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p"
)

const (
	OptimiseChannel = byte(0x50)


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

func (pR *P2PReactor) AddPeer(peer *p2p.Peer) {
// nothing to do yet
}

func (pR *P2PReactor) RemovePeer(peer *p2p.Peer, reason interface{}) {
// nothing to do yet
}

func (pR *P2PReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
// TODO
}

// implements events.Eventable
func (pR *P2PReactor) SetEventSwitch(evsw agtypes.EventSwitch) {
	pR.evsw = evsw
}









