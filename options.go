package bochka

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	timeout time.Duration
	image   string
	version string
	network *testcontainers.DockerNetwork
}

type option func(*options)

func getOptions(opts []option) (opt options) {
	opt.timeout = 30 * time.Second
	for _, o := range opts {
		o(&opt)
	}
	return
}

func WithCustomImage(image, version string) option {
	return func(opt *options) {
		opt.image = image
		opt.version = version
	}
}

func WithNetwork(network *testcontainers.DockerNetwork) option {
	return func(opt *options) {
		opt.network = network
	}
}
