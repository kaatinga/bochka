package bochka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func Test_setupPostgreDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start container with default port and run pgx query
	helper := New(t, ctx)
	err := helper.Start()
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	t.Logf("Started default port container on port %d", helper.Port())
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", helper.User(), helper.Password(), helper.Host(), helper.Port(), helper.DBName())
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		helper.Close()
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	_, err = conn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name TEXT)`)
	if err != nil {
		conn.Close(context.Background())
		helper.Close()
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = conn.Exec(context.Background(), `INSERT INTO test_table (name) VALUES ($1)`, "testname")
	if err != nil {
		conn.Close(context.Background())
		helper.Close()
		t.Fatalf("failed to insert row: %v", err)
	}
	var name string
	err = conn.QueryRow(context.Background(), `SELECT name FROM test_table WHERE name=$1`, "testname").Scan(&name)
	if err != nil {
		conn.Close(context.Background())
		helper.Close()
		t.Fatalf("failed to query row: %v", err)
	}
	if name != "testname" {
		conn.Close(context.Background())
		helper.Close()
		t.Errorf("expected name 'testname', got '%s'", name)
	}
	conn.Close(context.Background())
	helper.Close()

	// Start container with custom port and run pgx query
	helper = New(t, ctx, WithPort("5555"))
	err = helper.Start()
	if err != nil {
		t.Fatalf("failed to start container with custom port: %v", err)
	}
	t.Logf("Started custom port container on port %d", helper.Port())
	connStr = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", helper.User(), helper.Password(), helper.Host(), helper.Port(), helper.DBName())
	conn, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		helper.Close()
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	_, err = conn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS test_table2 (id SERIAL PRIMARY KEY, name TEXT)`)
	if err != nil {
		conn.Close(context.Background())
		helper.Close()
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = conn.Exec(context.Background(), `INSERT INTO test_table2 (name) VALUES ($1)`, "customport")
	if err != nil {
		conn.Close(context.Background())
		helper.Close()
		t.Fatalf("failed to insert row: %v", err)
	}
	err = conn.QueryRow(context.Background(), `SELECT name FROM test_table2 WHERE name=$1`, "customport").Scan(&name)
	if err != nil {
		conn.Close(context.Background())
		helper.Close()
		t.Fatalf("failed to query row: %v", err)
	}
	if name != "customport" {
		conn.Close(context.Background())
		helper.Close()
		t.Errorf("expected name 'customport', got '%s'", name)
	}
	conn.Close(context.Background())
	helper.Close()
}
