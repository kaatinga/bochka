# bochka for Golang

`bochka` is a Go package that streamlines your testing environment when working with Dockerized services. It provides helpers and primitives to initialize ready-to-use service containers within Docker, ideal for integration tests or any scenario where transient infrastructure is beneficial.

## Prerequisites

- **Go 1.25**
- **Docker** - Must be running and accessible
- **Docker Compose** (optional) - For more complex multi-service setups

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

## Quick Start

Here's a minimal example to get you started:

```go
package main

import (
	"context"
	"testing"
	"time"

	"github.com/kaatinga/bochka"
)

func TestQuickStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create and start a PostgreSQL container
	helper := bochka.NewPostgres(t, ctx)
	err := helper.Start()
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer helper.Close()

	// Container is ready! Use helper.Service().Host() and helper.Service().Port() to connect
	t.Logf("PostgreSQL running on %s:%d", helper.Service().Host(), helper.Service().Port())
}
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
		bochka.WithEnvVars(map[string]string{"POSTGRES_INITDB_ARGS": "--encoding=UTF8"}),
	)
	err := helper.Start()
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer helper.Close()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		helper.Service().User(), helper.Service().Password(), helper.Service().Host(), helper.Service().Port(), helper.Service().DBName())
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
	"fmt"
	"testing"
	"time"

	"github.com/kaatinga/bochka"
	"github.com/nats-io/nats.go"
)

func TestNatsContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	helper := bochka.NewNats(t, ctx,
		bochka.WithPort("4222"),
		bochka.WithCustomImage("docker.io/library/nats", "2-alpine"),
		bochka.WithEnvVars(map[string]string{"NATS_SERVER_NAME": "test-server"}),
	)
	err := helper.Start()
	if err != nil {
		t.Fatalf("failed to start NATS container: %v", err)
	}
	defer helper.Close()

	// Connect to NATS
	nc, err := nats.Connect(fmt.Sprintf("nats://%s:%d", helper.Service().Host(), helper.Service().Port()))
	if err != nil {
		t.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Your NATS operations here...
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
	defer network.Remove(ctx)

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
	if err := postgres.Start(); err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	defer postgres.Close()

	if err := nats.Start(); err != nil {
		t.Fatalf("failed to start nats: %v", err)
	}
	defer nats.Close()

	// Both containers can communicate via the shared network
}
```

## API

### Generic Container Management
- `func NewPostgres(t *testing.T, ctx context.Context, opts ...option) *Bochka[*PostgresService]`: Creates a new PostgreSQL test helper.
- `func NewNats(t *testing.T, ctx context.Context, opts ...option) *Bochka[*NatsService]`: Creates a new NATS test helper.
- `func (b *Bochka[T]) Start() error`: Starts the container.
- `func (b *Bochka[T]) Close() error`: Stops and removes the container.
- `func (b *Bochka[T]) NetworkName() string`: Returns the name of the Docker network used by the container.
- `func (b *Bochka[T]) Service() T`: Returns the underlying container service.
- `func (b *Bochka[T]) PrintLogs()`: Prints the container logs to the test output.

### PostgreSQL API
- `func (p *PostgresService) Host() string`: Returns the host address.
- `func (p *PostgresService) Port() uint16`: Returns the mapped port.
- `func (p *PostgresService) User() string`: Returns the username (default: "test").
- `func (p *PostgresService) Password() string`: Returns the password (default: "12345").
- `func (p *PostgresService) DBName() string`: Returns the database name (default: "testdb").
- `func (p *PostgresService) HostAlias() string`: Returns the network alias.

### NATS API
- `func (n *NatsService) Host() string`: Returns the host address.
- `func (n *NatsService) Port() uint16`: Returns the mapped port.
- `func (n *NatsService) HostAlias() string`: Returns the network alias.

### Options
- `WithPort(port string)`: Sets the host port for the container port binding.
- `WithCustomImage(image, version string)`: Sets a custom Docker image and version for the container.
- `WithNetwork(network *testcontainers.DockerNetwork)`: Sets a custom Docker network for the container to join.
- `WithEnvVars(vars map[string]string)`: Adds custom environment variables to the container. Multiple calls to `WithEnvVars` will merge the environment variables.

### Network Management
- `func NewNetwork(ctx context.Context) (*testcontainers.DockerNetwork, error)`: Creates a new Docker network for container communication.

## Extending for Other Services

The package uses a generic architecture with the `ContainerService` interface. To add a new service:

1. Implement the `ContainerService` interface:
```go
type ContainerService interface {
	Start(ctx context.Context) error
	Close() error
	HostAlias() string
	GetContainer() testcontainers.Container
}
```

2. Create a service struct and implement the interface methods. You may also add service-specific methods like `Host()`, `Port()`, `User()`, `Password()`, and `DBName()` as needed (though these are not required by the interface).

3. Add constructor and starter functions following the pattern:
```go
func NewYourService(t *testing.T, ctx context.Context, settings ...option) *Bochka[*YourService]
```

4. Add service-specific helper functions as needed.

## Troubleshooting

### Common Issues

**Docker not running**
```
Error: failed to start container: context deadline exceeded
```
**Solution**: Ensure Docker is running and accessible. Check with `docker ps`.

**Port already in use**
```
Error: failed to start container: port already allocated
```
**Solution**: Use `WithPort()` to specify a different port or ensure no other containers are using the same port.

**Image not found**
```
Error: failed to start container: image not found
```
**Solution**: Check the image name and version. Use `docker pull <image>:<version>` to pre-download the image.

**Context timeout**
```
Error: context deadline exceeded
```
**Solution**: Increase the context timeout or check if Docker has enough resources.

### Performance Tips

- Use specific image versions instead of `latest` for reproducible tests
- Reuse networks when running multiple services
- Set appropriate context timeouts based on your system's performance
- Consider using `WithCustomImage()` to use lighter Alpine-based images

## Comparison with Other Tools

| Feature | bochka | testcontainers-go | dockertest |
|---------|--------|-------------------|------------|
| Go Generics | ✅ | ❌ | ❌ |
| Type Safety | ✅ | ⚠️ | ❌ |
| Ease of Use | ✅ | ⚠️ | ✅ |
| Extensibility | ✅ | ✅ | ⚠️ |
| Network Support | ✅ | ✅ | ❌ |

## Contributing
If you would like to contribute to `bochka`, please raise an issue or submit a pull request on our GitHub repository.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
