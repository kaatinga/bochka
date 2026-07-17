package bochka

import (
	"context"
	"io"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

// ContainerService defines the interface that any container service must implement
type ContainerService interface {
	Start(ctx context.Context) error // Start is not supposed to be used. Use Bochka.Start()
	Close() error
	HostAlias() string
	GetContainer() testcontainers.Container
}

// ContainerConfig holds common configuration for any container
type ContainerConfig struct {
	EnvVars  map[string]string
	Files    map[string]string // container path -> content
	Cmd      []string
	Image    string
	Version  string
	HostPort string
}

// Bochka is a generic test helper for managing container lifecycles.
type Bochka[T ContainerService] struct {
	Context      context.Context
	t            *testing.T
	network      *testcontainers.DockerNetwork
	networkOwned bool // true if the network was created by the helper and must be removed on Close
	service      T
}

// NetworkName returns the name of the Docker network used by the container.
func (b *Bochka[T]) NetworkName() string {
	return b.network.Name
}

// Service returns the underlying container service
func (b *Bochka[T]) Service() T {
	return b.service
}

// Close terminates the container and removes the Docker network if it was
// created by the helper rather than supplied via WithNetwork.
func (b *Bochka[T]) Close() error {
	err := b.service.Close()
	if b.networkOwned && b.network != nil {
		if rmErr := b.network.Remove(context.Background()); rmErr != nil && err == nil {
			err = rmErr
		}
	}
	return err
}

func (b *Bochka[T]) Start() error {
	return b.Service().Start(b.Context)
}

// PrintLogs writes the container logs to the test output. Failures to fetch
// the logs are logged but do not fail the test.
func (b *Bochka[T]) PrintLogs() {
	logReader, err := b.Service().GetContainer().Logs(b.Context)
	if err != nil {
		b.t.Logf("failed to get %s container logs: %v", b.service.HostAlias(), err)
		return
	}

	defer func() {
		if err := logReader.Close(); err != nil {
			b.t.Logf("failed to close log reader: %v", err)
		}
	}()

	logs, err := io.ReadAll(logReader)
	if err != nil {
		b.t.Logf("failed to read %s container logs: %v", b.service.HostAlias(), err)
		return
	}

	b.t.Logf("%s container logs:\n%s", b.service.HostAlias(), logs)
}
