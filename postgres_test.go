package bochka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func Test_PostgreDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	// Start container with default port and run pgx query
	helper := NewPostgres(t, ctx)
	if err := helper.Start(); err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	t.Logf("Started default port container on port %d", helper.Service().Port())
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		helper.Service().User(),
		helper.Service().Password(),
		helper.Service().Host(),
		helper.Service().Port(),
		helper.Service().DBName(),
	)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	_, err = conn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name TEXT)`)
	if err != nil {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = conn.Exec(context.Background(), `INSERT INTO test_table (name) VALUES ($1)`, "testname")
	if err != nil {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to insert row: %v", err)
	}
	var name string
	err = conn.QueryRow(context.Background(), `SELECT name FROM test_table WHERE name=$1`, "testname").Scan(&name)
	if err != nil {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to query row: %v", err)
	}
	if name != "testname" {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Errorf("expected name 'testname', got '%s'", name)
	}
	if closeErr := conn.Close(context.Background()); closeErr != nil {
		t.Logf("failed to close conn: %v", closeErr)
	}
	if closeErr := helper.Close(); closeErr != nil {
		t.Logf("failed to close helper: %v", closeErr)
	}

	// Start container with custom port and run pgx query
	helper = NewPostgres(t, ctx, WithPort("5555"))
	err = helper.Service().Start(ctx)
	if err != nil {
		t.Fatalf("failed to start container with custom port: %v", err)
	}
	t.Logf("Started custom port container on port %d", helper.Service().Port())
	connStr = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", helper.Service().User(), helper.Service().Password(), helper.Service().Host(), helper.Service().Port(), helper.Service().DBName())
	conn, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	_, err = conn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_table2 (id SERIAL PRIMARY KEY, name TEXT)`)
	if err != nil {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = conn.Exec(context.Background(), `INSERT INTO test_table2 (name) VALUES ($1)`, "customport")
	if err != nil {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to insert row: %v", err)
	}
	err = conn.QueryRow(context.Background(), `SELECT name FROM test_table2 WHERE name=$1`, "customport").Scan(&name)
	if err != nil {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Fatalf("failed to query row: %v", err)
	}
	if name != "customport" {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			t.Logf("failed to close conn: %v", closeErr)
		}
		if closeErr := helper.Close(); closeErr != nil {
			t.Logf("failed to close helper: %v", closeErr)
		}
		t.Errorf("expected name 'customport', got '%s'", name)
	}
	if closeErr := conn.Close(context.Background()); closeErr != nil {
		t.Logf("failed to close conn: %v", closeErr)
	}
	if closeErr := helper.Close(); closeErr != nil {
		t.Logf("failed to close helper: %v", closeErr)
	}
}
