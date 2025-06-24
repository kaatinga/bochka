package bochka

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	faststrconv "github.com/kaatinga/strconv"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	login     = "test"
	password  = "12345"
	dbName    = "testdb"
	hostAlias = "postgres"
)

// Bochka is a test helper for managing a PostgreSQL container lifecycle.
type Bochka struct {
	Container testcontainers.Container
	context.Context
	options
	t *testing.T

	host    string
	port    uint16
	network *testcontainers.DockerNetwork
}

// HostAlias returns the network alias for the PostgreSQL container.
func (b *Bochka) HostAlias() string {
	return hostAlias
}

// NetworkName returns the name of the Docker network used by the container.
func (b *Bochka) NetworkName() string {
	return b.network.Name
}

// Host returns the host address of the PostgreSQL container.
func (b *Bochka) Host() string {
	return b.host
}

// Port returns the mapped port of the PostgreSQL container.
func (b *Bochka) Port() uint16 {
	return b.port
}

// User returns the username for the PostgreSQL instance.
func (b *Bochka) User() string {
	return login
}

// Password returns the password for the PostgreSQL instance.
func (b *Bochka) Password() string {
	return password
}

// DBName returns the database name for the PostgreSQL instance.
func (b *Bochka) DBName() string {
	return dbName
}

// Close terminates the PostgreSQL container.
func (b *Bochka) Close() error {
	return b.Container.Terminate(b.Context)
}

// New creates a new PostgreSQL test helper.
func New(t *testing.T, ctx context.Context, settings ...option) *Bochka {
	return &Bochka{
		t:       t,
		options: getOptions(settings),
		Context: ctx,
	}
}

// Start starts the PostgreSQL container and sets up connection details. Returns error on failure.
func (b *Bochka) Start() error {
	t := b.t
	t.Helper()

	if b.options.image == "" {
		b.options.image = "postgres"
	}
	if b.options.version == "" {
		b.options.version = "17.5"
	}

	if b.network == nil {
		var err error
		b.network, err = NewNetwork(b.Context)
		if err != nil {
			return err
		}
	}

	port := b.options.port
	if port == "" {
		port = "5433"
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        b.options.image + ":" + b.options.version,
		ExposedPorts: []string{"5432/tcp"},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
				"5432/tcp": {{HostIP: "", HostPort: port}},
			}
			hostConfig.AutoRemove = true
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     login,
			"POSTGRES_PASSWORD": password,
		},
		Networks: []string{b.network.Name},
		NetworkAliases: map[string][]string{
			b.network.Name: {hostAlias},
		},
	}

	t.Logf("Starting PostgreSQL container with version %s", b.options.version)

	var err error
	b.Container, err = testcontainers.GenericContainer(
		b.Context,
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return err
	}

	b.host, err = b.Container.Host(b.Context)
	if err != nil {
		return err
	}

	t.Logf("PostgreSQL host: %s", b.host)

	var mappedPort nat.Port
	mappedPort, err = b.Container.MappedPort(b.Context, "5432")
	if err != nil {
		return err
	}
	b.port, err = faststrconv.GetUint16(mappedPort.Port())
	if err != nil {
		return err
	}

	t.Logf("PostgreSQL port: %s", mappedPort.Port())
	return nil
}
