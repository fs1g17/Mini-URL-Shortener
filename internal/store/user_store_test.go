package store

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func getTestStore(t *testing.T) *UserStore {
	connString := "host=localhost user=postgres password=postgres dbname=test_db port=5432 sslmode=disable"

	err := Migrate(connString)
	if err != nil {
		t.Fatalf("Failed to run test migrations: %v\n", err)
	}

	pgConn := Connect("postgres://postgres:postgres@localhost:5432/test_db")
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

	user_id, err := userStore.CreateUser("theo", "password")
	if err != nil {
		t.Fatalf("didn't expect error, got: %v\n", err)
	}

	if user_id != 1 {
		t.Fatalf("user_id, got: %d, want: %d\n", user_id, 1)
	}

	// actually check db has this user
	var username string
	err = userStore.conn.QueryRow(context.Background(), "SELECT username FROM users WHERE id = $1;", user_id).Scan(&username)
	if err != nil {
		t.Fatalf("didn't expect error, got: %v\n", err)
	}
	if username != "theo" {
		t.Fatalf("got: %s, want: %s\n", username, "theo")
	}
}
