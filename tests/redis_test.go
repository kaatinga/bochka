package bochka_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kaatinga/bochka"
)

func Test_RedisService(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	helper := bochka.NewRedis(t, ctx)
	if err := helper.Start(); err != nil {
		t.Fatalf("failed to start Redis container: %v", err)
	}
	defer func() {
		if err := helper.Close(); err != nil {
			t.Logf("failed to close helper: %v", err)
		}
	}()

	svc := helper.Service()
	if svc.Host() == "" {
		t.Error("expected non-empty host")
	}
	if svc.Port() == 0 {
		t.Error("expected non-zero port")
	}
	if svc.HostAlias() != "redis" {
		t.Errorf("expected alias 'redis', got %q", svc.HostAlias())
	}
	wantAddr := fmt.Sprintf("%s:%d", svc.Host(), svc.Port())
	if svc.Addr() != wantAddr {
		t.Errorf("Addr(): got %q, want %q", svc.Addr(), wantAddr)
	}

	rdb := redis.NewClient(&redis.Options{Addr: svc.Addr()})
	defer func() { _ = rdb.Close() }()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis ping: %v", err)
	}
}

func Test_RedisWithCustomPort(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	helper := bochka.NewRedis(t, ctx, bochka.WithPort("6390"))
	if err := helper.Start(); err != nil {
		t.Fatalf("failed to start Redis container: %v", err)
	}
	defer func() {
		if err := helper.Close(); err != nil {
			t.Logf("failed to close helper: %v", err)
		}
	}()

	rdb := redis.NewClient(&redis.Options{Addr: helper.Service().Addr()})
	defer func() { _ = rdb.Close() }()

	if err := rdb.Set(ctx, "bochka-key", "ok", 0).Err(); err != nil {
		t.Fatalf("redis set: %v", err)
	}
	val, err := rdb.Get(ctx, "bochka-key").Result()
	if err != nil {
		t.Fatalf("redis get: %v", err)
	}
	if val != "ok" {
		t.Errorf("expected value 'ok', got %q", val)
	}
}
