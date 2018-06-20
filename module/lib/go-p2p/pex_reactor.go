// Copyright 2017 Baptist-Publication Information Technology Services Co.,Ltd.
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

package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"

	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p/discover"
	"github.com/Baptist-Publication/chorus/module/lib/go-wire"
)

var pexErrInvalidMessage = errors.New("Invalid PEX message")

const (
	PexChannel               = byte(0x00)
	ensurePeersPeriodSeconds = 30
	minNumOutboundPeers      = 10
	maxPexMessageSize        = 1048576 // 1MB
)

/*
PEXReactor handles PEX (peer exchange) and ensures that an
adequate number of peers are connected to the switch.
*/
type PEXReactor struct {
	BaseReactor

	discv   *discover.Network
	logger  *zap.Logger
	slogger *zap.SugaredLogger
}

func NewPEXReactor(logger *zap.Logger, discv *discover.Network) *PEXReactor {
	pexR := &PEXReactor{
		discv:   discv,
		logger:  logger,
		slogger: logger.Sugar(),
	}
	pexR.BaseReactor = *NewBaseReactor(logger, "PEXReactor", pexR)
	return pexR
}

func (pexR *PEXReactor) OnStart() error {
	pexR.BaseReactor.OnStart()
	go pexR.ensurePeersRoutine()
	return nil
}

func (pexR *PEXReactor) OnStop() {
	pexR.BaseReactor.OnStop()
}

// Implements Reactor
func (pexR *PEXReactor) GetChannels() []*ChannelDescriptor {
	return []*ChannelDescriptor{
		&ChannelDescriptor{
			ID:                PexChannel,
			Priority:          1,
			SendQueueCapacity: 10,
		},
	}
}

// Implements Reactor
func (pexR *PEXReactor) AddPeer(peer *Peer) {
	maxPeers := pexR.Switch.config.GetInt(configKeyMaxNumPeers)
	if pexR.Switch.Peers().Size() <= maxPeers {
		return
	}

	/*
		nodes := make([]*discover.Node, 10)
		if n := pexR.discv.ReadRandomNodes(nodes); n == 0 {
			return
		}

		if pexR.TrySendAddrs(peer, nodes) {
			<-time.After(1 * time.Second)
			pexR.Switch.StopPeerGracefully(peer)
		}
		pexR.logger.Warn("addPeer: reach the max peer, exchange then close")
	*/
	pexR.Switch.StopPeerGracefully(peer)
	pexR.logger.Warn("addPeer: reach the max peer, close connection")
	return
}

// Implements Reactor
func (pexR *PEXReactor) RemovePeer(peer *Peer, reason interface{}) {
	// TODO
}

// Implements Reactor
// Handles incoming PEX messages.
func (pexR *PEXReactor) Receive(chID byte, src *Peer, msgBytes []byte) {

	// decode message
	_, msg, err := DecodeMessage(msgBytes)
	if err != nil {
		pexR.logger.Warn("Error decoding message", zap.String("error", err.Error()))
		pexR.Switch.StopPeerGracefully(src)
		return
	}
	pexR.slogger.Infow("Received message", "msg", msg)

	switch msg := msg.(type) {
	case *pexRequestMessage:
		// src requested some peers.
		// TODO: prevent abuse.
		nodes := make([]*discover.Node, 10)
		if n := pexR.discv.ReadRandomNodes(nodes); n == 0 {
			return
		}
	case *pexAddrsMessage:
		// We received some peer addresses from src.
		// TODO: prevent abuse.
		// (We don't want to get spammed with bad peers)
		for _, nodeid := range msg.NodeIdSet {
			pexR.discv.Resolve(nodeid)
		}

	default:
		pexR.slogger.Warnf("Unknown message type %T", msg)
	}

}

// Asks peer for more addresses.
func (pexR *PEXReactor) RequestPEX(peer *Peer) {
	peer.Send(PexChannel, struct{ PexMessage }{&pexRequestMessage{}})
}

func (pexR *PEXReactor) SendAddrs(peer *Peer, nodes []*discover.Node) {
	nodeIdSet := []discover.NodeID{}
	for _, node := range nodes {
		if node == nil {
			break
		}
		nodeIdSet = append(nodeIdSet, node.ID)
	}
	peer.Send(PexChannel, struct{ PexMessage }{&pexAddrsMessage{NodeIdSet: nodeIdSet}})
}

// SendAddrs sends addrs to the peer.
func (r *PEXReactor) TrySendAddrs(p *Peer, nodes []*discover.Node) bool {
	nodeIdSet := []discover.NodeID{}
	for _, node := range nodes {
		if node == nil {
			break
		}
		nodeIdSet = append(nodeIdSet, node.ID)
	}

	ok := p.TrySend(PexChannel, struct{ PexMessage }{&pexAddrsMessage{NodeIdSet: nodeIdSet}})
	if !ok {
		r.Switch.StopPeerGracefully(p)
	}
	return ok
}

// Ensures that sufficient peers are connected. (continuous)
func (pexR *PEXReactor) ensurePeersRoutine() {
	// Randomize when routine starts
	time.Sleep(time.Duration(rand.Int63n(500*ensurePeersPeriodSeconds)) * time.Millisecond)

	// fire once immediately.
	pexR.ensurePeers()
	// fire periodically
	timer := NewRepeatTimer("pex", ensurePeersPeriodSeconds*time.Second)
FOR_LOOP:
	for {
		select {
		case <-timer.Ch:
			pexR.ensurePeers()
		case <-pexR.Quit:
			break FOR_LOOP
		}
	}

	// Cleanup
	timer.Stop()
}

func (pexR *PEXReactor) dialPeerWorker(a *NetAddress, wg *sync.WaitGroup) {
	if _, err := pexR.Switch.DialPeerWithAddress(a); err != nil {
		pexR.logger.Error("dialPeerWorker fail on dial peer", zap.String("addr ", a.String()), zap.String("err", err.Error()))
	}
	wg.Done()
}

func (pexR *PEXReactor) ensurePeers() {
	numOutPeers, _, numDialing := pexR.Switch.NumPeers()
	numToDial := (minNumOutboundPeers - (numOutPeers + numDialing))
	pexR.logger.Debug("Ensure peers",
		zap.Int("numOutPeers", numOutPeers),
		zap.Int("numDialing", numDialing),
		zap.Int("numToDial", numToDial))
	if numToDial <= 0 {
		return
	}

	connectedPeers := make(map[string]struct{})
	for _, peer := range pexR.Switch.Peers().List() {
		connectedPeers[peer.RemoteAddrHost()] = struct{}{}
	}

	toDial := make(map[string]*NetAddress)
	maxTry := numToDial * 3

	for i := 0; i < maxTry; i++ {
		nodes := make([]*discover.Node, 1)
		if n := pexR.discv.ReadRandomNodes(nodes); n != 1 {
			return
		}

		try := NewNetAddressIPPort(nodes[i].IP, nodes[i].TCP)
		if pexR.Switch.NodeInfo().ListenAddr == try.String() {
			continue
		}
		if dialling := pexR.Switch.IsDialing(try); dialling {
			continue
		}
		if _, ok := connectedPeers[try.IP.String()]; ok {
			continue
		}

		toDial[try.IP.String()] = try
	}

	var wg sync.WaitGroup
	for key, oneAttempt := range toDial {
		pexR.logger.Debug(fmt.Sprintf("Will dial address addr: %v", key))
		wg.Add(1)
		go pexR.dialPeerWorker(oneAttempt, &wg)
	}
	wg.Wait()
}

//-----------------------------------------------------------------------------
// Messages

const (
	msgTypeRequest = byte(0x01)
	msgTypeAddrs   = byte(0x02)
)

type PexMessage interface{}

var _ = wire.RegisterInterface(
	struct{ PexMessage }{},
	wire.ConcreteType{&pexRequestMessage{}, msgTypeRequest},
	wire.ConcreteType{&pexAddrsMessage{}, msgTypeAddrs},
)

func DecodeMessage(bz []byte) (msgType byte, msg PexMessage, err error) {
	msgType = bz[0]
	n := new(int)
	r := bytes.NewReader(bz)
	msg = wire.ReadBinary(struct{ PexMessage }{}, r, maxPexMessageSize, n, &err).(struct{ PexMessage }).PexMessage
	return
}

/*
A pexRequestMessage requests additional peer addresses.
*/
type pexRequestMessage struct {
}

func (m *pexRequestMessage) String() string {
	return "[pexRequest]"
}

/*
A message with announced peer addresses.
*/
type pexAddrsMessage struct {
	NodeIdSet []discover.NodeID
}

func (m *pexAddrsMessage) String() string {
	return fmt.Sprintf("[pexAddrs %v]", m.NodeIdSet)
}
