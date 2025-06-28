# bochka for Golang

`bochka` is a Go package that streamlines your testing environment when working with Dockerized services. It provides helpers and primitives to initialize ready-to-use service containers within Docker, ideal for integration tests or any scenario where transient infrastructure is beneficial.

## Features
- **Ephemeral Service Instances**: Quickly set up service containers (e.g., PostgreSQL, NATS, Redis, etc.) that last only for the duration of your tests.
- **Generic Architecture**: Extensible design using Go generics and interfaces for easy addition of new services.
- **Version & Image Control**: Easily specify the desired version and image of any service.
- **Custom Port Support**: Run multiple containers in parallel or avoid port conflicts by specifying the host port.
- **Environment Variables**: Configure services with custom environment variables.
- **Shared Networks**: Run multiple services on the same Docker network for inter-service communication.
- **Seamless Docker Integration**: No need for complex Docker setups—`bochka` handles it for you.

## Installation

```bash
go get -u github.com/kaatinga/bochka
```

## Supported Services

### PostgreSQL
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

	helper := bochka.NewPostgres(t, ctx, 
		bochka.WithPort("5555"), 
		bochka.WithCustomImage("postgres", "17.5"),
	)
	err := bochka.StartPostgres(helper)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer helper.Close()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		bochka.User(helper), bochka.Password(helper), bochka.Host(helper), bochka.Port(helper), bochka.DBName(helper))
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer conn.Close(ctx)

	// Your database operations here...
}
```

### NATS
```go
package bochka_test

import (
	"context"
	"testing"
	"time"

	"github.com/kaatinga/bochka"
)

func TestNatsContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	helper := bochka.NewNats(t, ctx,
		bochka.WithPort("4222"),
		bochka.WithCustomImage("docker.io/library/nats", "2-alpine"),
		bochka.WithEnvVar("NATS_SERVER_NAME", "test-server"),
	)
	err := bochka.StartNats(helper)
	if err != nil {
		t.Fatalf("failed to start NATS container: %v", err)
	}
	defer helper.Close()

	// Connect to NATS using bochka.NatsHost(helper) and bochka.NatsPort(helper)
}
```

### Multiple Services on Shared Network
```go
func TestMultipleServices(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a shared network
	network, err := bochka.NewNetwork(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}

	// Create PostgreSQL container
	postgres := bochka.NewPostgres(t, ctx,
		bochka.WithNetwork(network),
		bochka.WithPort("5433"),
	)

	// Create NATS container on the same network
	nats := bochka.NewNats(t, ctx,
		bochka.WithNetwork(network),
		bochka.WithPort("4222"),
	)

	// Start both containers
	if err := bochka.StartPostgres(postgres); err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	defer postgres.Close()

	if err := bochka.StartNats(nats); err != nil {
		t.Fatalf("failed to start nats: %v", err)
	}
	defer nats.Close()

	// Both containers can communicate via the shared network
}
```

## API

### Generic Container Management
- `func NewPostgres(t *testing.T, ctx context.Context, opts ...option) *Bochka[*PostgresService]`
- `func NewNats(t *testing.T, ctx context.Context, opts ...option) *Bochka[*NatsService]`
- `func (b *Bochka[T]) Close() error`: Stops and removes the container.
- `func (b *Bochka[T]) NetworkName() string`: Returns the name of the Docker network.

### PostgreSQL API
- `func StartPostgres(b *Bochka[*PostgresService]) error`: Starts the PostgreSQL container.
- `func Host(b *Bochka[*PostgresService]) string`: Returns the host address.
- `func Port(b *Bochka[*PostgresService]) uint16`: Returns the mapped port.
- `func User(b *Bochka[*PostgresService]) string`: Returns the username (default: "test").
- `func Password(b *Bochka[*PostgresService]) string`: Returns the password (default: "12345").
- `func DBName(b *Bochka[*PostgresService]) string`: Returns the database name (default: "testdb").
- `func HostAlias(b *Bochka[*PostgresService]) string`: Returns the network alias.

### NATS API
- `func StartNats(b *Bochka[*NatsService]) error`: Starts the NATS container.
- `func NatsHost(b *Bochka[*NatsService]) string`: Returns the host address.
- `func NatsPort(b *Bochka[*NatsService]) uint16`: Returns the mapped port.
- `func NatsHostAlias(b *Bochka[*NatsService]) string`: Returns the network alias.

### Options
- `WithPort(port string)`: Set the host port for the service.
- `WithCustomImage(image, version string)`: Set the Docker image and version.
- `WithNetwork(network *testcontainers.DockerNetwork)`: Attach to a custom Docker network.
- `WithEnvVar(key, value string)`: Add custom environment variables.

### Network Management
- `func NewNetwork(ctx context.Context) (*testcontainers.DockerNetwork, error)`: Creates a new Docker network.

## Extending for Other Services

The package uses a generic architecture with the `ContainerService` interface. To add a new service:

1. Implement the `ContainerService` interface:
```go
type ContainerService interface {
	Start(ctx context.Context) error
	Close() error
	NetworkName() string
	Host() string
	Port() uint16
}
```

2. Create a service struct and implement the interface methods.

3. Add constructor and starter functions following the pattern:
```go
func NewYourService(t *testing.T, ctx context.Context, settings ...option) *Bochka[*YourService]
func StartYourService(b *Bochka[*YourService]) error
```

4. Add service-specific helper functions as needed.

## Contributing
If you would like to contribute to `bochka`, please raise an issue or submit a pull request on our GitHub repository.
