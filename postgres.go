package bochka

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	faststrconv "github.com/kaatinga/strconv"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	login     = "test"
	password  = "12345"
	dbName    = "testdb"
	hostAlias = "postgres"
)

// PostgresService implements ContainerService for PostgreSQL
type PostgresService struct {
	Container testcontainers.Container
	host      string
	port      uint16
	network   *testcontainers.DockerNetwork
	config    ContainerConfig
}

// Start starts the PostgreSQL container and sets up connection details. Returns error on failure.
func (p *PostgresService) Start(ctx context.Context) error {
	envVars := map[string]string{
		"POSTGRES_DB":       dbName,
		"POSTGRES_USER":     login,
		"POSTGRES_PASSWORD": password,
	}

	for env, val := range p.config.EnvVars {
		envVars[env] = val
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        p.config.Image + ":" + p.config.Version,
		ExposedPorts: []string{"5432/tcp"},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
				"5432/tcp": {{HostIP: "", HostPort: faststrconv.Uint162String(p.Port())}},
			}
			hostConfig.AutoRemove = true
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
		Env:      envVars,
		Networks: []string{p.network.Name},
		NetworkAliases: map[string][]string{
			p.network.Name: {hostAlias},
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

	var mappedPort nat.Port
	mappedPort, err = p.Container.MappedPort(ctx, "5432")
	if err != nil {
		return err
	}
	p.port, err = faststrconv.GetUint16(mappedPort.Port())
	if err != nil {
		return err
	}

	return nil
}

// Close terminates the PostgreSQL container.
func (p *PostgresService) Close() error {
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
	return hostAlias
}

// User returns the username for the PostgreSQL instance.
func (p *PostgresService) User() string {
	return login
}

// Password returns the password for the PostgreSQL instance.
func (p *PostgresService) Password() string {
	return password
}

// DBName returns the database name for the PostgreSQL instance.
func (p *PostgresService) DBName() string {
	return dbName
}

// NewPostgres creates a new PostgreSQL test helper.
func NewPostgres(t *testing.T, ctx context.Context, settings ...option) *Bochka[*PostgresService] {
	opts := options{
		// default settings
		image:   "postgres",
		version: "17.5",
		port:    "5433",
	}

	opts.applyOptions(settings)

	network := opts.network
	if network == nil {
		var err error
		network, err = NewNetwork(ctx)
		if err != nil {
			t.Fatalf("failed to create network: %v", err)
		}
	}

	service := &PostgresService{
		network: network,
		config: ContainerConfig{
			Image:    opts.image,
			Version:  opts.version,
			HostPort: opts.port,
		},
	}

	bochka := &Bochka[*PostgresService]{
		t:       t,
		options: opts,
		Context: ctx,
		network: network,
		service: service,
	}

	return bochka
}
