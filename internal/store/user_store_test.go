package store

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func getTestStore(t *testing.T) *UserStore {
	connString := "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable"

	err := Migrate(connString)
	if err != nil {
		t.Fatalf("Failed to run test migrations: %v\n", err)
	}

	pgConn := Connect("postgres://postgres:postgres@localhost:5433/postgres")
	userStore := NewUserStore(pgConn)
	return userStore
}

func TestCreateUser(t *testing.T) {
	userStore := getTestStore(t)

	//ensure no users present
	var userCount int
	err := userStore.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM users;").Scan(&userCount)
	if err != nil {
		t.Fatalf("got error querying user count: %v\n", err)
	}
	if userCount != 0 {
		t.Fatalf("expected 0 users but got %d\n", userCount)
	}

}
