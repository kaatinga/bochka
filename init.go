package bochka

import (
	"net/netip"

	"github.com/moby/moby/api/types/network"
)

// AnyIP binds container ports on all host interfaces.
var AnyIP = netip.IPv4Unspecified()

// mustParsePort parses a TCP port number into a network.Port and panics on
// failure. It is used only for compile-time constant ports.
func mustParsePort(port string) network.Port {
	p, err := network.ParsePort(port + "/tcp")
	if err != nil {
		panic(err)
	}
	return p
}
