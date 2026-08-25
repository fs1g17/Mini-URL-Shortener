package store

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func Connect() *pgx.Conn {
	// pgx v5 takes a connection string rather than a ConnConfig literal.
	const connString = "postgres://postgres:postgres@localhost:5432/postgres"

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
