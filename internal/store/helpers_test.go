package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/fs1g17/Mini-URL-Shortener/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func migrate(connString string) error {
	testDb, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("opening %s db: %v", connString, err)
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(testDb, "."); err != nil {
		return err
	}

	return nil
}

func connectToTestDb(t *testing.T) *pgx.Conn {
	connString := "host=localhost user=postgres password=postgres dbname=test_db port=5432 sslmode=disable"

	err := migrate(connString)
	if err != nil {
		t.Fatalf("Failed to run test migrations: %v\n", err)
	}

	pgConn := Connect("postgres://postgres:postgres@localhost:5432/test_db")
	return pgConn
}

func checkTestDb(pgConn *pgx.Conn, t *testing.T) {
	var db string
	pgConn.QueryRow(context.Background(), "SELECT current_database()").Scan(&db)

	if db != "test_db" {
		t.Fatalf("want: test_db, got: %s\n", db)
	}
}

func resetDataBase(pgConn *pgx.Conn, t *testing.T) {
	_, err := pgConn.Exec(context.Background(), "TRUNCATE users, links, click_events RESTART IDENTITY;")

	if err != nil {
		t.Fatalf("Failed to truncate table: %v\n", err)
	}
}

func getTestUserStore(t *testing.T) *UserStore {
	pgConn := connectToTestDb(t)
	checkTestDb(pgConn, t)
	resetDataBase(pgConn, t)
	userStore := NewUserStore(pgConn)
	return userStore
}

func getTestLinkStore(t *testing.T) *LinkStore {
	pgConn := connectToTestDb(t)
	checkTestDb(pgConn, t)
	resetDataBase(pgConn, t)
	linkStore := NewLinkStore(pgConn)
	return linkStore
}
