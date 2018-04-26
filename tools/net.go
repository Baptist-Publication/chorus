package tools

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/Baptist-Publication/chorus/module/lib/go-p2p/upnp"
)

func ProtocolAndAddress(listenAddr string) (string, string) {
	protocol, address := "tcp", listenAddr
	parts := strings.SplitN(address, "://", 2)
	if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	}
	return protocol, address
}

func DetermineExternalAddress(listener net.Listener, lAddrIP, lAddrPort string, skipUPNP bool) (string, error) {
	var err error
	externalPort, err := strconv.Atoi(lAddrPort)
	if err != nil {
		return "", err
	}
	_, listenerPort, _ := net.SplitHostPort(listener.Addr().String())
	internalPort, _ := strconv.Atoi(listenerPort)

	external := ""
	if !skipUPNP {
		if lAddrIP == "" || lAddrIP == "0.0.0.0" {
			var nat upnp.NAT
			var ext net.IP
			nat, err = upnp.Discover()
			if err != nil {
				goto NATIVE
			}
			ext, err = nat.GetExternalAddress()
			if err != nil {
				goto NATIVE
			}
			externalPort, err = nat.AddPortMapping("tcp", externalPort, internalPort, "chorus", 0)
			if err != nil {
				goto NATIVE
			}
			external = fmt.Sprintf("%s:%d", ext.String(), externalPort)
		}
	}

NATIVE:
	if external == "" {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return "", errors.Wrap(err, "Could not fetch interface addresses")
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 {
				continue
			}

			external = fmt.Sprintf("%s:%s", v4.String(), listenerPort)
			break
		}
	}

	return external, nil
}
