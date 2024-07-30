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
)

type PostgreTestHelper struct {
	Container  testcontainers.Container
	Pool       *pgxpool.Pool
	CancelFunc context.CancelFunc
	context.Context
	options
	t *testing.T
}

func (helper *PostgreTestHelper) Close() {
	helper.CancelFunc()
	if helper.Pool != nil {
		helper.Pool.Close()
	}

	_ = helper.Container.Terminate(context.Background())
}

// NewPostgreTestHelper creates a new PostgreSQL test helper.
func NewPostgreTestHelper(t *testing.T, settings ...option) *PostgreTestHelper {
	helper := &PostgreTestHelper{
		t:       t,
		options: getOptions(settings),
	}
	helper.Context, helper.CancelFunc = context.WithTimeout(context.Background(), helper.timeout)

	return helper
}

// Run starts PostgreSQL container and creates a connection pool. The version parameter is used to specify the
// PostgreSQL version. The version must be in the format of "major.minor", e.g. "14.5".
func (helper *PostgreTestHelper) Run(version string) {
	t := helper.t
	t.Helper()

	// 1. Create PostgreSQL container request.
	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:" + version,
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     login,
			"POSTGRES_PASSWORD": password,
		},
	}

	// 2. Start PostgreSQL container.
	var err error
	helper.Container, err = testcontainers.GenericContainer(
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
	host, err = helper.Container.Host(helper.Context)
	if err != nil {
		t.Fatal(err)
	}

	var port nat.Port
	port, err = helper.Container.MappedPort(helper.Context, "5432")
	if err != nil {
		t.Fatal(err)
	}

	// 3.2 Create DB connection string and connect.
	connectionURI := fmt.Sprintf("postgres://%s:%s@%v:%v/testdb", login, password, host, port.Port())
	helper.Pool, err = pgxpool.New(helper.Context, connectionURI)
	if err != nil {
		t.Fatal(err)
	}
}
