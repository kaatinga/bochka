package bochka

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

func NewNetwork(ctx context.Context) (*testcontainers.DockerNetwork, error) {
	dockerNetwork, err := network.New(ctx, network.WithAttachable())
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	return dockerNetwork, nil
}
