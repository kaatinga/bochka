package postgreSQLtesthelper

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v4/pgxpool"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgreDatabase(login, password string, t *testing.T) (testcontainers.Container, *pgxpool.Pool) {
	ctx := context.Background()

	// 1. Create PostgreSQL container request.
	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:14.5",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     login,
			"POSTGRES_PASSWORD": password,
		},
	}

	// 2. Start PostgreSQL container.
	dbContainer, _ := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})

	// 3.1 Get host and port of PostgreSQL container.
	host, err := dbContainer.Host(ctx)
	if err != nil {
		t.Error(err)
		return nil, nil
	}

	var port nat.Port
	port, err = dbContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Error(err)
		return nil, nil
	}

	// 3.2 Create db connection string and connect.
	dbURI := fmt.Sprintf("postgres://%s:%s@%v:%v/testdb", login, password, host, port.Port())
	var connPool *pgxpool.Pool
	connPool, err = pgxpool.Connect(ctx, dbURI)
	if err != nil {
		t.Error(err)
		return nil, nil
	}

	return dbContainer, connPool
}
