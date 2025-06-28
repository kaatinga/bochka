package bochka

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	timeout      time.Duration
	image        string
	version      string
	network      *testcontainers.DockerNetwork
	port         string // Host port for container
	extraEnvVars map[string]string
}

type option func(*options)

func (o *options) applyOptions(opts []option) {
	o.timeout = 30 * time.Second
	o.extraEnvVars = make(map[string]string)
	for _, opt := range opts {
		opt(o)
	}
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

func WithPort(port string) option {
	return func(opt *options) {
		opt.port = port
	}
}

func WithEnvVars(vars map[string]string) option {
	return func(opt *options) {
		for k, v := range vars {
			opt.extraEnvVars[k] = v
		}
	}
}
