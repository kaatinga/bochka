# bochka for Golang

`bochka` is a Go package that streamlines your testing environment when working with Dockerized services. It provides helpers and primitives to initialize ready-to-use service containers within Docker, ideal for integration tests or any scenario where transient infrastructure is beneficial.

## Features
- **Ephemeral Service Instances**: Quickly set up service containers (e.g., PostgreSQL, Redis, etc.) that last only for the duration of your tests.
- **Version & Image Control**: Easily specify the desired version and image of any service.
- **Custom Port Support**: Run multiple containers in parallel or avoid port conflicts by specifying the host port.
- **Seamless Docker Integration**: No need for complex Docker setups—`bochka` handles it for you.
- **Extensible**: Designed to be extended for any service, not just databases.

## Installation

```bash
go get -u github.com/kaatinga/bochka
```

## Example Usage (PostgreSQL)

Below is an example using the current API and [pgx/v5](https://github.com/jackc/pgx) for database access. You can use similar patterns for other services by extending `bochka`:

```go
package bochka_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kaatinga/bochka"
	"github.com/jackc/pgx/v5"
)

func TestPostgresContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	helper := bochka.New(t, ctx, bochka.WithPort("5555"), bochka.WithCustomImage("postgres", "17.5"))
	err := helper.Start()
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
    defer helper.Close()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		helper.User(), helper.Password(), helper.Host(), helper.Port(), helper.DBName())
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO test_table (name) VALUES ($1)`, "testname")
	if err != nil {
		t.Fatalf("failed to insert row: %v", err)
	}
	var name string
	err = conn.QueryRow(ctx, `SELECT name FROM test_table WHERE name=$1`, "testname").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}
	if name != "testname" {
		t.Errorf("expected name 'testname', got '%s'", name)
	}
}
```

## API

### `func New(t *testing.T, ctx context.Context, opts ...option) *Bochka`
Creates a new test helper for a service container. Use options to customize the container.

### `func (b *Bochka) Start() error`
Starts the service container. Returns an error if startup fails.

### `func (b *Bochka) Close() error`
Stops and removes the container.

### Options
- `WithPort(port string)`: Set the host port for the service (default: "5433" for Postgres).
- `WithCustomImage(image, version string)`: Set the Docker image and version.
- `WithNetwork(network *testcontainers.DockerNetwork)`: Attach to a custom Docker network.

### Helper Methods (PostgreSQL)
- `Host() string`: Returns the host address of the PostgreSQL container.
- `Port() uint16`: Returns the mapped port of the PostgreSQL container.
- `User() string`: Returns the username for the PostgreSQL instance (default: "test").
- `Password() string`: Returns the password for the PostgreSQL instance (default: "12345").
- `DBName() string`: Returns the database name for the PostgreSQL instance (default: "testdb").
- `HostAlias() string`: Returns the network alias for the PostgreSQL container.
- `NetworkName() string`: Returns the name of the Docker network used by the container.

## Extending for Other Services

While the current implementation provides helpers for PostgreSQL, you can extend `bochka` to support other services (e.g., Redis, MySQL, etc.) by following similar patterns. Contributions are welcome!

## Contributing
If you would like to contribute to `bochka`, please raise an issue or submit a pull request on our GitHub repository.
