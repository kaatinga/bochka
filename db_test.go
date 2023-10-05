package postgreSQLtesthelper

import (
	"context"
	"testing"
)

func TestSetupPostgreDatabase(t *testing.T) {
	t.Run("run_container", func(t *testing.T) {
		helper := SetupPostgreTestHelper(t, "14.5")
		defer helper.Close()
		err := helper.Pool.Ping(context.Background())
		if err != nil {
			t.Error("ping failed:", err)
		}
	})
}
