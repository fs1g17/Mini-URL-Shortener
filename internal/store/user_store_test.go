package store

import (
	"context"
	"testing"
)

func TestCreateUser(t *testing.T) {
	userStore := getTestUserStore(t)

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
