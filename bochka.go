package bochka

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

// ContainerService defines the interface that any container service must implement
type ContainerService interface {
	Start(ctx context.Context) error
	Close() error
	NetworkName() string
	Host() string
	Port() uint16
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
