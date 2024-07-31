package bochka

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	login    = "test"
	password = "12345"
	dbName   = "testdb"
)

type Bochka struct {
	Container  testcontainers.Container
	CancelFunc context.CancelFunc
	context.Context
	options
	t *testing.T

	connectionURI string
}

func (b *Bochka) ConnectionURI() string {
	return b.connectionURI
}

func (b *Bochka) Close() {
	b.CancelFunc()

	_ = b.Container.Terminate(b.Context)
}

// New creates a new PostgreSQL test helper.
func New(t *testing.T, ctx context.Context, settings ...option) *Bochka {
	helper := &Bochka{
		t:       t,
		options: getOptions(settings),
	}
	helper.Context, helper.CancelFunc = context.WithTimeout(ctx, helper.timeout)

	return helper
}

// Run starts PostgreSQL container and creates a connection pool. The version parameter is used to specify the
// PostgreSQL version. The version must be in the format of "major.minor", e.g. "14.5".
func (b *Bochka) Run(version string) {
	t := b.t
	t.Helper()

	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:16.3", // Specify the PostgreSQL version as needed
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "12345",
		},
	}

	t.Logf("Starting PostgreSQL container with version %s", version)

	// 2. Start PostgreSQL container.
	var err error
	b.Container, err = testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("PostgreSQL container started with ID %s", b.Container.GetContainerID())

	// 3.1 Get host and port of PostgreSQL container.
	var host string
	host, err = b.Container.Host(b.Context)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("PostgreSQL host: %s", host)

	var port nat.Port
	port, err = b.Container.MappedPort(b.Context, "5432")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("PostgreSQL port: %s", port.Port())

	b.connectionURI = fmt.Sprintf("postgres://%s:%s@%v:%v/%s", login, password, host, port.Port(), dbName)
}
