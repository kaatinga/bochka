package bochka_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kaatinga/bochka"
)

func Test_PostgreDatabase(t *testing.T) {
	run := func(t *testing.T, opts ...bochka.Option) {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer cancel()

		helper := bochka.NewPostgres(t, ctx, opts...)
		if err := helper.Start(); err != nil {
			t.Fatalf("failed to start container: %v", err)
		}
		t.Cleanup(func() {
			if err := helper.Close(); err != nil {
				t.Logf("failed to close helper: %v", err)
			}
		})
		t.Logf("Started container on port %d", helper.Service().Port())

		conn, err := pgx.Connect(ctx, helper.Service().DSN())
		if err != nil {
			t.Fatalf("failed to connect to postgres: %v", err)
		}
		t.Cleanup(func() {
			if err := conn.Close(context.Background()); err != nil {
				t.Logf("failed to close conn: %v", err)
			}
		})

		_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name TEXT)`)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
		_, err = conn.Exec(ctx, `INSERT INTO test_table (name) VALUES ($1)`, "testname")
		if err != nil {
			t.Fatalf("failed to insert row: %v", err)
		}
		var name string
		err = conn.QueryRow(ctx, `SELECT name FROM test_table WHERE name=$1`, "testname").Scan(&name)
		if err != nil {
			t.Fatalf("failed to query row: %v", err)
		}
		if name != "testname" {
			t.Errorf("expected name 'testname', got '%s'", name)
		}
	}

	t.Run("default port", func(t *testing.T) {
		run(t)
	})

	t.Run("custom port", func(t *testing.T) {
		run(t, bochka.WithPort("5555"))
	})
}
