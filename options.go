package bochka

import (
	"maps"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	network      *testcontainers.DockerNetwork
	extraEnvVars map[string]string
	files        map[string]string // container path -> file content
	cmd          []string
	image        string
	version      string
	port         string // Host port for container
}

// Option configures a container helper created by NewPostgres, NewRedis or NewNats.
type Option func(*options)

func (o *options) applyOptions(opts []Option) {
	o.extraEnvVars = make(map[string]string)
	o.files = make(map[string]string)
	for _, opt := range opts {
		opt(o)
	}
}

// WithCustomImage sets a custom Docker image and version for the container.
func WithCustomImage(image, version string) Option {
	return func(opt *options) {
		opt.image = image
		opt.version = version
	}
}

// WithNetwork sets a custom Docker network for the container to join.
func WithNetwork(network *testcontainers.DockerNetwork) Option {
	return func(opt *options) {
		opt.network = network
	}
}

// WithPort sets the host port for the container port binding.
// An empty string lets Docker pick a random free port; the actual port is
// available via the service's Port() method after Start.
func WithPort(port string) Option {
	return func(opt *options) {
		opt.port = port
	}
}

// WithEnvVars adds custom environment variables to the container.
// Multiple calls to WithEnvVars will merge the environment variables.
func WithEnvVars(vars map[string]string) Option {
	return func(opt *options) {
		maps.Copy(opt.extraEnvVars, vars)
	}
}

// WithCmd overrides the container command. When unset, each service uses its default
// (for NATS: nats-server -js).
func WithCmd(args ...string) Option {
	return func(opt *options) {
		opt.cmd = append([]string(nil), args...)
	}
}

// WithFiles mounts files into the container. Keys are absolute container paths;
// values are the file contents written at those paths. Multiple calls merge entries.
func WithFiles(files map[string]string) Option {
	return func(opt *options) {
		maps.Copy(opt.files, files)
	}
}
