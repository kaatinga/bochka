package bochka

import (
	"context"
	"testing"
	"time"
)

func TestSetupPostgreDatabase(t *testing.T) {
	t.Run("run_container", func(t *testing.T) {
		helper := New(t, WithTimeout(10*time.Second))
		helper.Run("14.5")
		defer helper.Close()
		err := helper.Pool.Ping(context.Background())
		if err != nil {
			t.Error("ping failed:", err)
		}
	})
}
