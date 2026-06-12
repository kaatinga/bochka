package bochka

import (
	"context"
	"maps"
	"testing"

	faststrconv "github.com/kaatinga/strconv"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresLogin     = "test"
	postgresPassword  = "12345"
	postgresDBName    = "testdb"
	postgresHostAlias = "postgres"
	postgresPort      = "5432"
)

var postgresExposedPort = mustParsePort(postgresPort)

// PostgresService implements ContainerService for PostgreSQL
type PostgresService struct {
	Container testcontainers.Container
	network   *testcontainers.DockerNetwork
	config    ContainerConfig
	host      string
	port      uint16
}

// Start starts the PostgreSQL container and sets up connection details. Returns error on failure.
func (p *PostgresService) Start(ctx context.Context) error {
	envVars := map[string]string{
		"POSTGRES_DB":       postgresDBName,
		"POSTGRES_USER":     postgresLogin,
		"POSTGRES_PASSWORD": postgresPassword,
	}

	maps.Copy(envVars, p.config.EnvVars)

	containerReq := testcontainers.ContainerRequest{
		Image:        p.config.Image + ":" + p.config.Version,
		ExposedPorts: []string{postgresExposedPort.String()},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// No AutoRemove: it races with Terminate ("removal already in
			// progress"); the reaper cleans up if the test crashes.
			hostConfig.PortBindings = network.PortMap{
				postgresExposedPort: {{HostIP: AnyIP, HostPort: p.config.HostPort}},
			}
		},
		WaitingFor: wait.ForAll(
			// Postgres emits this line twice: once from the temporary server
			// started by initdb and once from the real one.
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort(postgresExposedPort.String()),
		),
		Env:      envVars,
		Networks: []string{p.network.Name},
		NetworkAliases: map[string][]string{
			p.network.Name: {postgresHostAlias},
		},
	}

	var err error
	p.Container, err = testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return err
	}

	p.host, err = p.Container.Host(ctx)
	if err != nil {
		return err
	}

	mappedPort, err := p.Container.MappedPort(ctx, postgresPort)
	if err != nil {
		return err
	}
	p.port, err = faststrconv.GetUint16(mappedPort.Port())
	if err != nil {
		return err
	}

	return nil
}

// Close terminates the PostgreSQL container. It is safe to call even if the
// container was never started.
func (p *PostgresService) Close() error {
	if p.Container == nil {
		return nil
	}
	return p.Container.Terminate(context.Background())
}

// NetworkName returns the name of the Docker network used by the container.
func (p *PostgresService) NetworkName() string {
	return p.network.Name
}

// Host returns the host address of the PostgreSQL container.
func (p *PostgresService) Host() string {
	return p.host
}

// Port returns the mapped port of the PostgreSQL container.
func (p *PostgresService) Port() uint16 {
	return p.port
}

// HostAlias returns the network alias for the PostgreSQL container.
func (p *PostgresService) HostAlias() string {
	return postgresHostAlias
}

// User returns the username for the PostgreSQL instance.
func (p *PostgresService) User() string {
	return postgresLogin
}

// Password returns the password for the PostgreSQL instance.
func (p *PostgresService) Password() string {
	return postgresPassword
}

// DBName returns the database name for the PostgreSQL instance.
func (p *PostgresService) DBName() string {
	return postgresDBName
}

// GetContainer returns the underlying container service
func (p *PostgresService) GetContainer() testcontainers.Container {
	return p.Container
}

// DSN returns a connection string for the PostgreSQL instance.
func (p *PostgresService) DSN() string {
	return "postgres://" + p.User() + ":" + p.Password() + "@" + p.Host() + ":" + faststrconv.Uint162String(p.Port()) + "/" + p.DBName() + "?sslmode=disable"
}

// NewPostgres creates a new PostgreSQL test helper.
func NewPostgres(t *testing.T, ctx context.Context, settings ...Option) *Bochka[*PostgresService] {
	opts := options{
		// default settings
		image:   "postgres",
		version: "17.5",
		port:    "5433",
	}

	opts.applyOptions(settings)

	network := opts.network
	networkOwned := false
	if network == nil {
		var err error
		network, err = NewNetwork(ctx)
		if err != nil {
			t.Fatalf("failed to create network: %v", err)
		}
		networkOwned = true
	}

	service := &PostgresService{
		network: network,
		config: ContainerConfig{
			Image:    opts.image,
			Version:  opts.version,
			HostPort: opts.port,
			EnvVars:  opts.extraEnvVars,
		},
	}

	bochka := &Bochka[*PostgresService]{
		t:            t,
		Context:      ctx,
		network:      network,
		networkOwned: networkOwned,
		service:      service,
	}

	return bochka
}
