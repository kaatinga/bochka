package bochka

import (
	"context"
	"testing"
	"time"
)

func Test_setupPostgreDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	helper := New(t, ctx)
	t.Cleanup(func() {
		cancel()
		helper.Close()
	})
	t.Run("run_container", func(t *testing.T) {
		helper.Run()
	})
}
