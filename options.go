package bochka

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	timeout time.Duration
	image   string
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

func WithCustomImage(image string) option {
	return func(opt *options) {
		opt.image = image
	}
}

func WithNetwork(network *testcontainers.DockerNetwork) option {
	return func(opt *options) {
		opt.network = network
	}
}
