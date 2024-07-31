package bochka

import (
	"context"
	"testing"
	"time"
)

func TestSetupPostgreDatabase(t *testing.T) {
	helper := New(t, context.Background(), WithTimeout(10*time.Second))
	t.Cleanup(func() {
		helper.Close()
	})
	t.Run("run_container", func(t *testing.T) {
		helper.Run("14.5")
	})
}
