package bochka

import (
	"context"
	"testing"

	faststrconv "github.com/kaatinga/strconv"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	redisHostAlias = "redis"
	redisPort      = "6379"
)

var (
	redisExposedPort network.Port
)

// RedisService implements ContainerService for Redis.
type RedisService struct {
	Container testcontainers.Container
	network   *testcontainers.DockerNetwork
	config    ContainerConfig
	host      string
	port      uint16
}

// Start starts the Redis container and sets up connection details. Returns error on failure.
func (r *RedisService) Start(ctx context.Context) error {
	envVars := r.config.EnvVars
	if envVars == nil {
		envVars = make(map[string]string)
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        r.config.Image + ":" + r.config.Version,
		ExposedPorts: []string{redisExposedPort.String()},
		Env:          envVars,
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort(redisExposedPort.String()),
		),
		Networks: []string{r.network.Name},
		NetworkAliases: map[string][]string{
			r.network.Name: {redisHostAlias},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = network.PortMap{
				redisExposedPort: {
					{
						HostIP:   AnyIP,
						HostPort: r.config.HostPort,
					},
				},
			}
			hostConfig.AutoRemove = true
		},
	}

	var err error
	r.Container, err = testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return err
	}

	r.host, err = r.Container.Host(ctx)
	if err != nil {
		return err
	}

	mappedPort, err := r.Container.MappedPort(ctx, redisPort)
	if err != nil {
		return err
	}

	r.port, err = faststrconv.GetUint16(mappedPort.Port())
	if err != nil {
		return err
	}

	return nil
}

// Close terminates the Redis container.
func (r *RedisService) Close() error {
	return r.Container.Terminate(context.Background())
}

// NetworkName returns the name of the Docker network used by the container.
func (r *RedisService) NetworkName() string {
	return r.network.Name
}

// Host returns the host address of the Redis container.
func (r *RedisService) Host() string {
	return r.host
}

// Port returns the mapped port of the Redis container.
func (r *RedisService) Port() uint16 {
	return r.port
}

// HostAlias returns the network alias for the Redis container.
func (r *RedisService) HostAlias() string {
	return redisHostAlias
}

// GetContainer returns the underlying container service.
func (r *RedisService) GetContainer() testcontainers.Container {
	return r.Container
}

// Addr returns host:port for Redis connections.
func (r *RedisService) Addr() string {
	return r.Host() + ":" + faststrconv.Uint162String(r.Port())
}

// NewRedis creates a new Redis test helper.
func NewRedis(t *testing.T, ctx context.Context, settings ...option) *Bochka[*RedisService] {
	opts := options{
		image:   "redis",
		version: "7-alpine",
		port:    "6380",
	}

	opts.applyOptions(settings)

	net := opts.network
	if net == nil {
		var err error
		net, err = NewNetwork(ctx)
		if err != nil {
			t.Fatalf("failed to create network: %v", err)
		}
	}

	service := &RedisService{
		network: net,
		config: ContainerConfig{
			Image:    opts.image,
			Version:  opts.version,
			HostPort: opts.port,
			EnvVars:  opts.extraEnvVars,
		},
	}

	b := &Bochka[*RedisService]{
		t:       t,
		options: opts,
		Context: ctx,
		network: net,
		service: service,
	}

	return b
}
