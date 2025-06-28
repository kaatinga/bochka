package bochka

import (
	"context"
	"io"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

// ContainerService defines the interface that any container service must implement
type ContainerService interface {
	Start(ctx context.Context) error // Start is not supposed to be used. Use bochka.Start()
	Close() error
	NetworkName() string
	Host() string
	Port() uint16
	HostAlias() string
	User() string
	Password() string
	DBName() string
	GetContainer() testcontainers.Container
}

// ContainerConfig holds common configuration for any container
type ContainerConfig struct {
	Image        string
	Version      string
	ExposedPorts []string
	EnvVars      map[string]string
	NetworkAlias string
	HostPort     string
}

// Bochka is a generic test helper for managing container lifecycles.
type Bochka[T ContainerService] struct {
	Context context.Context
	options
	t       *testing.T
	network *testcontainers.DockerNetwork
	service T
}

// NetworkName returns the name of the Docker network used by the container.
func (b *Bochka[T]) NetworkName() string {
	return b.network.Name
}

// Service returns the underlying container service
func (b *Bochka[T]) Service() T {
	return b.service
}

// Close terminates the container
func (b *Bochka[T]) Close() error {
	return b.service.Close()
}

func (b *Bochka[T]) Start() error {
	return b.Service().Start(b.Context)
}

func (b *Bochka[T]) PrintLogs() {
	logReader, err := b.Service().GetContainer().Logs(b.Context)
	if err != nil {
		b.t.Errorf("failed to get %s container logs: %v", b.service.HostAlias(), err)
		return
	}

	defer logReader.Close()

	logs, err := io.ReadAll(logReader)
	if err != nil {
		b.t.Errorf("failed to get %s container logs: %v", b.service.HostAlias(), err)
		return
	}

	b.t.Logf("%s container logs:\n%s", b.service.HostAlias(), logs)
}
