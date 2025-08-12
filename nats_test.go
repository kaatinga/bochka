package bochka

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func Test_NatsService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start NATS container
	helper := NewNats(t, ctx, WithPort("4222"))
	if err := helper.Start(); err != nil {
		t.Fatalf("failed to start NATS container: %v", err)
	}
	defer helper.Close()

	// Verify connection details
	host := helper.Service().Host()
	port := helper.Service().Port()
	alias := helper.Service().HostAlias()

	if host == "" {
		t.Error("expected non-empty host")
	}
	if port == 0 {
		t.Error("expected non-zero port")
	}
	if alias != "nats" {
		t.Errorf("expected alias 'nats', got '%s'", alias)
	}

	// Connect to NATS server
	natsURL := "nats://" + host + ":" + strconv.Itoa(int(port))
	nc, err := nats.Connect(natsURL)
	if err != nil {
		t.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	if !nc.IsConnected() {
		t.Error("NATS client is not connected")
	}

	t.Logf("NATS container started successfully")
	t.Logf("NATS connection: %s:%d", host, port)
	t.Logf("NATS network alias: %s", alias)
}

func TestNatsWithCustomEnvVars(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start NATS container with custom environment variables
	helper := NewNats(t, ctx,
		WithPort("4223"),
		WithEnvVars(map[string]string{"NATS_SERVER_NAME": "test-server"}),
		WithEnvVars(map[string]string{"NATS_CLUSTER_NAME": "test-cluster"}),
	)
	if err := helper.Start(); err != nil {
		t.Fatalf("failed to start NATS container: %v", err)
	}
	defer helper.Close()

	t.Logf("NATS container started with custom environment variables")
	t.Logf("NATS connection: %s:%d", helper.Service().Host(), helper.Service().Port())
}
