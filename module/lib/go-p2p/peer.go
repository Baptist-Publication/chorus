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
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-wire"
	"github.com/Baptist-Publication/chorus/module/xlib"
)

type (
	Peer struct {
		BaseService

		outbound bool
		mconn    *MConnection

		*NodeInfo
		Key  string
		Data *CMap // User data.
	}

	AuthorizationFunc func(nodeinfo *NodeInfo) error
)

// NOTE: blocking
// Before creating a peer with newPeer(), perform a handshake on connection.
func peerHandshake(conn net.Conn, sw *Switch) (*NodeInfo, error) {
	var (
		peerNodeInfo = new(NodeInfo)
		err1         error
		err2         error
	)
	Parallel(
		func() {
			var bys []byte
			if bys, err1 = json.Marshal(sw.nodeInfo); err1 != nil {
				return
			}
			if err1 = xlib.WriteBytes(conn, bys); err1 != nil {
				return
			}
		},
		func() {
			var bys []byte
			if bys, err2 = xlib.ReadBytes(conn); err2 != nil {
				return
			}
			err2 = json.Unmarshal(bys, peerNodeInfo)
		})
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	if err := sw.AuthByCA(peerNodeInfo); err != nil {
		return nil, err
	}
	peerNodeInfo.RemoteAddr = conn.RemoteAddr().String()
	return peerNodeInfo, nil
}

// NOTE: call peerHandshake on conn before calling newPeer().
func newPeer(logger *zap.Logger, config *viper.Viper, conn net.Conn, peerNodeInfo *NodeInfo, outbound bool, reactorsByCh map[byte]Reactor, chDescs []*ChannelDescriptor, onPeerError func(*Peer, interface{})) *Peer {
	var p *Peer
	onReceive := func(chID byte, msgBytes []byte) {
		reactor := reactorsByCh[chID]
		if reactor == nil {
			PanicSanity(Fmt("Unknown channel %X", chID))
		}
		reactor.Receive(chID, p, msgBytes)
	}
	onError := func(r interface{}) {
		p.Stop()
		onPeerError(p, r)
	}
	mconn := NewMConnection(logger, config, conn, chDescs, onReceive, onError)
	p = &Peer{
		outbound: outbound,
		mconn:    mconn,
		NodeInfo: peerNodeInfo,
		Key:      peerNodeInfo.PubKey.KeyString(),
		Data:     NewCMap(),
	}
	p.BaseService = *NewBaseService(logger, "Peer", p)
	return p
}

func (p *Peer) OnStart() error {
	p.BaseService.OnStart()
	_, err := p.mconn.Start()
	return err
}

func (p *Peer) OnStop() {
	p.BaseService.OnStop()
	p.mconn.Stop()
}

func (p *Peer) Connection() *MConnection {
	return p.mconn
}

func (p *Peer) IsOutbound() bool {
	return p.outbound
}

func (p *Peer) Send(chID byte, msg interface{}) bool {
	return p.SendBytes(chID, wire.BinaryBytes(msg))
}

func (p *Peer) SendBytes(chID byte, msg []byte) bool {
	if !p.IsRunning() {
		return false
	}
	return p.mconn.Send(chID, msg)
}

func (p *Peer) TrySend(chID byte, msg interface{}) bool {
	return p.TrySendBytes(chID, wire.BinaryBytes(msg))
}

func (p *Peer) TrySendBytes(chID byte, msg []byte) bool {
	if !p.IsRunning() {
		return false
	}
	return p.mconn.TrySend(chID, msg)
}

func (p *Peer) CanSend(chID byte) bool {
	if !p.IsRunning() {
		return false
	}
	return p.mconn.CanSend(chID)
}

func (p *Peer) WriteTo(w io.Writer) (n int64, err error) {
	var n_ int
	wire.WriteString(p.Key, w, &n_, &err)
	n += int64(n_)
	return
}

func (p *Peer) String() string {
	if p.outbound {
		return fmt.Sprintf("Peer{%v %v out}", p.mconn, p.Key[:12])
	} else {
		return fmt.Sprintf("Peer{%v %v in}", p.mconn, p.Key[:12])
	}
}

func (p *Peer) Equals(other *Peer) bool {
	return p.Key == other.Key
}

func (p *Peer) Get(key string) interface{} {
	return p.Data.Get(key)
}
