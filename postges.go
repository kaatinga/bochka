package bochka

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Pool       *pgxpool.Pool
	CancelFunc context.CancelFunc
	context.Context
	options
	t *testing.T
}

func (b *Bochka) Close() {
	b.CancelFunc()
	if b.Pool != nil {
		b.Pool.Close()
	}

	_ = b.Container.Terminate(context.Background())
}

func (b *Bochka) Port() string {
	return b.port
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

	if b.port == "" {
		b.port = "5432"
	}

	// 1. Create PostgreSQL container request.
	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:" + version,
		ExposedPorts: []string{b.port + "/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort(nat.Port(b.port+"/tcp")),
		),
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     login,
			"POSTGRES_PASSWORD": password,
		},
	}

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

	// 3.1 Get host and port of PostgreSQL container.
	var host string
	host, err = b.Container.Host(b.Context)
	if err != nil {
		t.Fatal(err)
	}

	var port nat.Port
	port, err = b.Container.MappedPort(b.Context, nat.Port(b.port))
	if err != nil {
		t.Fatal(err)
	}

	// 3.2 Create DB connection string and connect.
	connectionURI := fmt.Sprintf("postgres://%s:%s@%v:%v/%s", login, password, host, port.Port(), dbName)
	b.Pool, err = pgxpool.New(b.Context, connectionURI)
	if err != nil {
		t.Fatal(err)
	}
}
