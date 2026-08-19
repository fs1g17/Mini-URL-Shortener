package store

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx"
)

func Connect() *pgx.Conn {
	var config pgx.ConnConfig = pgx.ConnConfig{
		User:     "postgres",
		Password: "postgres",
		Host:     "localhost",
		Port:     5432,
	}

	conn, err := pgx.Connect(config)
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
