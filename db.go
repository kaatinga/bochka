package postgreSQLtesthelper

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v4/pgxpool"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type postgreTestHelper struct {
	Container  testcontainers.Container
	Pool       *pgxpool.Pool
	CancelFunc context.CancelFunc
	context.Context
}

func (h *postgreTestHelper) Close() {
	h.CancelFunc()
	h.Pool.Close()
	_ = h.Container.Terminate(context.Background())
}

const (
	login    = "test"
	password = "12345"
)

func SetupPostgreTestHelper(t *testing.T, version string) *postgreTestHelper {
	t.Helper()

	var helper postgreTestHelper
	helper.Context, helper.CancelFunc = context.WithTimeout(context.Background(), 30*time.Second)

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
		return nil
	}

	// 3.1 Get host and port of PostgreSQL container.
	var host string
	host, err = helper.Container.Host(helper.Context)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	var port nat.Port
	port, err = helper.Container.MappedPort(helper.Context, "5432")
	if err != nil {
		t.Fatal(err)
		return nil
	}

	// 3.2 Create DB connection string and connect.
	connectionURI := fmt.Sprintf("postgres://%s:%s@%v:%v/testdb", login, password, host, port.Port())
	helper.Pool, err = pgxpool.Connect(helper.Context, connectionURI)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	return &helper
}
