package bochka

import (
	"net/netip"

	"github.com/moby/moby/api/types/network"
)

var (
	AnyIP netip.Addr
)

func init() {
	natsExposedPort, _ = network.ParsePort(natsPort + "/tcp")
	postgresExposedPort, _ = network.ParsePort(postgresPort + "/tcp")
	redisExposedPort, _ = network.ParsePort(redisPort + "/tcp")
	AnyIP, _ = netip.ParseAddr("0.0.0.0")
}
