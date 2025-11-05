package bochka

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	network      *testcontainers.DockerNetwork
	extraEnvVars map[string]string
	image        string
	version      string
	port         string // Host port for container
}

type option func(*options)

func (o *options) applyOptions(opts []option) {
	o.extraEnvVars = make(map[string]string)
	for _, opt := range opts {
		opt(o)
	}
}

// WithCustomImage sets a custom Docker image and version for the container.
func WithCustomImage(image, version string) option {
	return func(opt *options) {
		opt.image = image
		opt.version = version
	}
}

// WithNetwork sets a custom Docker network for the container to join.
func WithNetwork(network *testcontainers.DockerNetwork) option {
	return func(opt *options) {
		opt.network = network
	}
}

// WithPort sets the host port for the container port binding.
func WithPort(port string) option {
	return func(opt *options) {
		opt.port = port
	}
}

// WithEnvVars adds custom environment variables to the container.
// Multiple calls to WithEnvVars will merge the environment variables.
func WithEnvVars(vars map[string]string) option {
	return func(opt *options) {
		for k, v := range vars {
			opt.extraEnvVars[k] = v
		}
	}
}
