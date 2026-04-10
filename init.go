package bochka

import (
	"log"
	"net/netip"

	"github.com/moby/moby/api/types/network"
)

var (
	AnyIP netip.Addr
)

func init() {
	natsExposedPort, _ = network.ParsePort(natsPort + "/tcp")
	postgresExposedPort, _ = network.ParsePort(postgresPort + "/tcp")
	AnyIP, _ = netip.ParseAddr("0.0.0.0")
	log.Println(natsExposedPort)
	log.Println(postgresExposedPort)
	log.Println(AnyIP)
}
