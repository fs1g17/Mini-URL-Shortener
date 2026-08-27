package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fs1g17/Mini-URL-Shortener/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
)

func Connect(connString string) *pgx.Conn {
	// pgx v5 takes a connection string rather than a ConnConfig literal.
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v\n", err)
		os.Exit(1)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatalf("Failed to ping postgres: %v\n", err)
	}

	return conn
}

func Migrate(connString string) error {
	testDb, err := sql.Open("pgx", connString)
	if err != nil {
		return errors.New(fmt.Sprintf("opening %s db: %v", connString, err))
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
