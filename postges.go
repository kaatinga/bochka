package bochka

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/kaatinga/strconv"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	login     = "test"
	password  = "12345"
	dbName    = "testdb"
	hostAlias = "postgres"
)

type Bochka struct {
	Container testcontainers.Container
	context.Context
	options
	t *testing.T

	host    string
	port    uint16
	network *testcontainers.DockerNetwork
}

func (b *Bochka) HostAlias() string {
	return hostAlias
}

func (b *Bochka) NetworkName() string {
	return b.network.Name
}

func (b *Bochka) Host() string {
	return b.host
}

func (b *Bochka) Port() uint16 {
	return b.port
}

func (b *Bochka) User() string {
	return login
}

func (b *Bochka) Password() string {
	return password
}

func (b *Bochka) DBName() string {
	return dbName
}

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

// Run starts PostgreSQL container and creates a connection pool.
func (b *Bochka) Run() {
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
			t.Fatal(err)
		}
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        b.options.image + ":" + b.options.version,
		ExposedPorts: []string{"5432/tcp"},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
				"5432/tcp": {{HostIP: "", HostPort: "5433"}},
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
		t.Fatal(err)
	}

	// t.Logf("PostgreSQL container started with ID %s", b.Container.GetContainerID())

	b.host, err = b.Container.Host(b.Context)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("PostgreSQL host: %s", b.host)

	var port nat.Port
	port, err = b.Container.MappedPort(b.Context, "5432")
	if err != nil {
		t.Fatal(err)
	}
	b.port, err = faststrconv.GetUint16(port.Port())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("PostgreSQL port: %s", port.Port())
}
