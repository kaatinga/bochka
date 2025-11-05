package bochka

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	faststrconv "github.com/kaatinga/strconv"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	natsHostAlias   = "nats"
	natsPort        = "4222"
	natsExposedPort = "4222/tcp"
)

// NatsService implements ContainerService for NATS
type NatsService struct {
	Container testcontainers.Container
	network   *testcontainers.DockerNetwork
	config    ContainerConfig
	host      string
	port      uint16
}

// Start starts the NATS container and sets up connection details. Returns error on failure.
func (n *NatsService) Start(ctx context.Context) error {
	envVars := n.config.EnvVars
	if envVars == nil {
		envVars = make(map[string]string)
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        n.config.Image + ":" + n.config.Version,
		Cmd:          []string{"nats-server", "-js"},
		ExposedPorts: []string{natsExposedPort},
		Env:          envVars,
		WaitingFor: wait.ForAll(
			wait.ForLog("Server is ready").WithStartupTimeout(30*time.Second),
			wait.ForListeningPort(natsExposedPort),
		),
		Networks: []string{n.network.Name},
		NetworkAliases: map[string][]string{
			n.network.Name: {natsHostAlias},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
				natsExposedPort: {
					{
						HostIP:   "0.0.0.0",
						HostPort: n.config.HostPort,
					},
				},
			}
			hostConfig.AutoRemove = false // to see logs
		},
	}

	var err error
	n.Container, err = testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return err
	}

	n.host, err = n.Container.Host(ctx)
	if err != nil {
		return err
	}

	var mappedPort nat.Port
	mappedPort, err = n.Container.MappedPort(ctx, natsPort)
	if err != nil {
		return err
	}

	n.port, err = faststrconv.GetUint16(mappedPort.Port())
	if err != nil {
		return err
	}

	return nil
}

// Close terminates the NATS container.
func (n *NatsService) Close() error {
	return n.Container.Terminate(context.Background())
}

// NetworkName returns the name of the Docker network used by the container.
func (n *NatsService) NetworkName() string {
	return n.network.Name
}

// Host returns the host address of the NATS container.
func (n *NatsService) Host() string {
	return n.host
}

// Port returns the mapped port of the NATS container.
func (n *NatsService) Port() uint16 {
	return n.port
}

// HostAlias returns the network alias for the NATS container.
func (n *NatsService) HostAlias() string {
	return natsHostAlias
}

// GetContainer returns the underlying container service
func (n *NatsService) GetContainer() testcontainers.Container {
	return n.Container
}

// NewNats creates a new NATS test helper.
func NewNats(t *testing.T, ctx context.Context, settings ...option) *Bochka[*NatsService] {
	opts := options{
		// default settings
		image:   "docker.io/library/nats",
		version: "2-alpine",
		port:    natsPort,
	}

	opts.applyOptions(settings)

	network := opts.network
	if network == nil {
		var err error
		network, err = NewNetwork(ctx)
		if err != nil {
			t.Fatalf("failed to create network: %v", err)
		}
	}

	service := &NatsService{
		network: network,
		config: ContainerConfig{
			Image:    opts.image,
			Version:  opts.version,
			HostPort: opts.port,
			EnvVars:  opts.extraEnvVars,
		},
	}

	bochka := &Bochka[*NatsService]{
		t:       t,
		options: opts,
		Context: ctx,
		network: network,
		service: service,
	}

	return bochka
}
