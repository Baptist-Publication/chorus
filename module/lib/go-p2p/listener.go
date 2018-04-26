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
	"fmt"
	"net"
	"strconv"
	"time"

	"go.uber.org/zap"

	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p/upnp"
)

type Listener interface {
	Connections() <-chan net.Conn
	InternalAddress() *NetAddress
	ExternalAddress() *NetAddress
	String() string
	Stop() bool
}

// Implements Listener
type DefaultListener struct {
	BaseService

	listener    net.Listener
	intAddr     *NetAddress
	extAddr     *NetAddress
	connections chan net.Conn

	logger  *zap.Logger
	slogger *zap.SugaredLogger
}

const (
	numBufferedConnections = 10
	defaultExternalPort    = 8770
	tryListenSeconds       = 5
)

func splitHostPort(addr string) (host string, port int) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		PanicSanity(err)
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		PanicSanity(err)
	}
	return host, port
}

// skipUPNP: If true, does not try getUPNPExternalAddress()
func NewDefaultListener(logger *zap.Logger, protocol string, lAddr string, skipUPNP bool) (Listener, error) {
	// Local listen IP & port
	lAddrIP, lAddrPort := splitHostPort(lAddr)

	// Create listener
	var listener net.Listener
	var err error
	for i := 0; i < tryListenSeconds; i++ {
		listener, err = net.Listen(protocol, lAddr)
		if err == nil {
			break
		} else if i < tryListenSeconds-1 {
			time.Sleep(time.Second * 1)
		}
	}
	if err != nil {
		logger.Error("fail to create listener", zap.Error(err))
		return nil, err
	}
	// Actual listener local IP & port
	_, listenerPort := splitHostPort(listener.Addr().String())

	// Determine internal address...
	var intAddr *NetAddress
	intAddr, err = NewNetAddressString(lAddr)
	if err != nil {
		PanicCrisis(err)
	}

	// Determine external address...
	var extAddr *NetAddress
	if lAddrIP != "" && lAddrIP != "0.0.0.0" && lAddrIP != "127.0.0.1" {
		// Prefer address from config
		extAddr = intAddr
	} else if !skipUPNP {
		// Then prefer UPnP
		extAddr = getUPNPExternalAddress(logger, lAddrPort, listenerPort)
	} else {
		// At last we determine it from system network config
		nAddrs := getNaiveExternalAddress(listenerPort)
		for _, v := range nAddrs {
			if isPublicIP(v.IP) {
				extAddr = v
				break
			}
		}
		if extAddr == nil && len(nAddrs) > 0 {
			extAddr = nAddrs[0]
		}
	}

	if extAddr == nil {
		logger.Error("Could not determine external address!")
		return nil, fmt.Errorf("could not determine external address")
	}

	dl := &DefaultListener{
		listener:    listener,
		intAddr:     intAddr,
		extAddr:     extAddr,
		connections: make(chan net.Conn, numBufferedConnections),
		logger:      logger,
		slogger:     logger.Sugar(),
	}
	dl.BaseService = *NewBaseService(logger, "DefaultListener", dl)
	dl.Start() // Started upon construction
	return dl, nil
}

func isPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func (l *DefaultListener) OnStart() error {
	l.BaseService.OnStart()
	go l.listenRoutine()
	return nil
}

func (l *DefaultListener) OnStop() {
	l.BaseService.OnStop()
	l.listener.Close()
}

// Accept connections and pass on the channel
func (l *DefaultListener) listenRoutine() {
	for {
		conn, err := l.listener.Accept()

		if !l.IsRunning() {
			break // Go to cleanup
		}

		// listener wasn't stopped,
		// yet we encountered an error.
		if err != nil {
			PanicCrisis(err)
		}

		l.connections <- conn
	}

	// Cleanup
	close(l.connections)
	for _ = range l.connections {
		// Drain
	}
}

// A channel of inbound connections.
// It gets closed when the listener closes.
func (l *DefaultListener) Connections() <-chan net.Conn {
	return l.connections
}

func (l *DefaultListener) InternalAddress() *NetAddress {
	return l.intAddr
}

func (l *DefaultListener) ExternalAddress() *NetAddress {
	return l.extAddr
}

// NOTE: The returned listener is already Accept()'ing.
// So it's not suitable to pass into http.Serve().
func (l *DefaultListener) NetListener() net.Listener {
	return l.listener
}

func (l *DefaultListener) String() string {
	return fmt.Sprintf("Listener(@%v)", l.extAddr)
}

/* external address helpers */

// UPNP external address discovery & port mapping
func getUPNPExternalAddress(logger *zap.Logger, externalPort, internalPort int) *NetAddress {
	logger.Debug("Getting UPNP external address")
	nat, err := upnp.Discover()
	if err != nil {
		logger.Debug("Could not perform UPNP discover", zap.String("error", err.Error()))
		return nil
	}

	ext, err := nat.GetExternalAddress()
	if err != nil {
		logger.Debug("Could not get UPNP external address", zap.String("error", err.Error()))
		return nil
	}

	// UPnP can't seem to get the external port, so let's just be explicit.
	if externalPort == 0 {
		externalPort = defaultExternalPort
	}

	externalPort, err = nat.AddPortMapping("tcp", externalPort, internalPort, "tendermint", 0)
	if err != nil {
		logger.Debug("Could not add UPNP port mapping", zap.String("error", err.Error()))
		return nil
	}

	logger.Debug("Got UPNP external address", zap.Stringer("address", ext))
	return NewNetAddressIPPort(ext, uint16(externalPort))
}

// TODO: use syscalls: http://pastebin.com/9exZG4rh
func getNaiveExternalAddress(port int) []*NetAddress {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		PanicCrisis(Fmt("Could not fetch interface addresses: %v", err))
	}

	naddrs := make([]*NetAddress, 0)
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		v4 := ipnet.IP.To4()
		if v4 == nil || v4[0] == 127 {
			continue
		} // loopback
		naddrs = append(naddrs, NewNetAddressIPPort(ipnet.IP, uint16(port)))
	}
	return naddrs
}
