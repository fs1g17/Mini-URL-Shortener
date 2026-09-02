package store

import (
	"context"
	"errors"
	"testing"
)

func TestCreateUser(t *testing.T) {
	t.Run("should create new user", func(t *testing.T) {
		userStore := getTestUserStore(t)
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
	})

	t.Run("should fail creating duplicate user", func(t *testing.T) {
		userStore := getTestUserStore(t)
		_, err := userStore.CreateUser("theo", "password")
		if err != nil {
			t.Fatalf("didn't expect error, got: %v\n", err)
		}

		_, err = userStore.CreateUser("theo", "password")
		if !errors.Is(err, UsernameTakenErr) {
			t.Fatalf("want: username taken error, got: %v\n", err)
		}
	})
}

func TestSignIn(t *testing.T) {
	t.Run("should sign in with existing user", func(t *testing.T) {
		userStore := getTestUserStore(t)
		_, err := userStore.CreateUser("theo", "password")
		if err != nil {
			t.Fatalf("didn't expect error, got: %v\n", err)
		}

		user_id, err := userStore.SignIn("theo", "password")
		if err != nil {
			t.Fatalf("didn't epxect error, got: %v\n", err)
		}

		if user_id != 1 {
			t.Fatalf("want: user_id = 1, got: %d\n", user_id)
		}
	})

	t.Run("should fail to sign in with wrong password", func(t *testing.T) {
		userStore := getTestUserStore(t)
		_, err := userStore.SignIn("theo", "wrong_password")
		if !errors.Is(err, IncorrectInfoErr) {
			t.Fatalf("want: incorrect info error, got: %v\n", err)
		}
	})

	t.Run("should fail to sign in with wrong password", func(t *testing.T) {
		userStore := getTestUserStore(t)
		_, err := userStore.SignIn("cleo", "wrong_password")
		if !errors.Is(err, IncorrectInfoErr) {
			t.Fatalf("want: incorrect info error, got: %v\n", err)
		}
	})
}
