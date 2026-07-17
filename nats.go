package bochka

import (
	"context"
	"strings"
	"testing"
	"time"

	faststrconv "github.com/kaatinga/strconv"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	natsHostAlias = "nats"
	natsPort      = "4222"
)

var natsExposedPort = mustParsePort(natsPort)

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

	cmd := n.config.Cmd
	if len(cmd) == 0 {
		cmd = []string{"nats-server", "-js"}
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        n.config.Image + ":" + n.config.Version,
		Cmd:          cmd,
		ExposedPorts: []string{natsExposedPort.String()},
		Env:          envVars,
		Files:        containerFiles(n.config.Files),
		WaitingFor: wait.ForAll(
			wait.ForLog("Server is ready").WithStartupTimeout(30*time.Second),
			wait.ForListeningPort(natsExposedPort.String()),
		),
		Networks: []string{n.network.Name},
		NetworkAliases: map[string][]string{
			n.network.Name: {natsHostAlias},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = network.PortMap{
				natsExposedPort: {
					{
						HostIP:   AnyIP,
						HostPort: n.config.HostPort,
					},
				},
			}
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

	mappedPort, err := n.Container.MappedPort(ctx, natsPort)
	if err != nil {
		return err
	}

	n.port, err = faststrconv.GetUint16(mappedPort.Port())
	if err != nil {
		return err
	}

	return nil
}

// Close terminates the NATS container. It is safe to call even if the
// container was never started.
func (n *NatsService) Close() error {
	if n.Container == nil {
		return nil
	}
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

// Addr returns host:port for NATS client connections.
func (n *NatsService) Addr() string {
	return n.host + ":" + faststrconv.Uint162String(n.port)
}

// GetContainer returns the underlying container service
func (n *NatsService) GetContainer() testcontainers.Container {
	return n.Container
}

func containerFiles(files map[string]string) []testcontainers.ContainerFile {
	if len(files) == 0 {
		return nil
	}
	out := make([]testcontainers.ContainerFile, 0, len(files))
	for path, content := range files {
		out = append(out, testcontainers.ContainerFile{
			Reader:            strings.NewReader(content),
			ContainerFilePath: path,
			FileMode:          0o644,
		})
	}
	return out
}

// NewNats creates a new NATS test helper.
func NewNats(t *testing.T, ctx context.Context, settings ...Option) *Bochka[*NatsService] {
	opts := options{
		// default settings
		image:   "docker.io/library/nats",
		version: "2-alpine",
		// not the standard 4222 to avoid clashing with a locally running NATS
		port: "14222",
	}

	opts.applyOptions(settings)

	network := opts.network
	networkOwned := false
	if network == nil {
		var err error
		network, err = NewNetwork(ctx)
		if err != nil {
			t.Fatalf("failed to create network: %v", err)
		}
		networkOwned = true
	}

	service := &NatsService{
		network: network,
		config: ContainerConfig{
			Image:    opts.image,
			Version:  opts.version,
			HostPort: opts.port,
			EnvVars:  opts.extraEnvVars,
			Cmd:      opts.cmd,
			Files:    opts.files,
		},
	}

	bochka := &Bochka[*NatsService]{
		t:            t,
		Context:      ctx,
		network:      network,
		networkOwned: networkOwned,
		service:      service,
	}

	return bochka
}
