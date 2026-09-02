package store

import (
	"context"
	"testing"
)

func TestCreateShortenedURL(t *testing.T) {

	t.Run("create shortened url with no expiration for existing user", func(t *testing.T) {
		linkStore := getTestLinkStore(t)
		userStore := getTestUserStore(t)

		user_id, err := userStore.CreateUser("theo", "password")
		if err != nil {
			t.Fatalf("expected no error, got: %v\n", err)
		}

		slug, err := linkStore.CreateShortenedURL("https://google.com", user_id, ShortUrlConfig{})
		if err != nil {
			t.Fatalf("expected no error, got: %v\n", err)
		}

		var redirect_url string
		var click_count int
		var click_limit *int
		var expires *string
		var owner_id int

		// verify slug is in the database
		err = linkStore.conn.QueryRow(
			context.Background(),
			"SELECT redirect_url,click_count,click_limit,expires,owner_id FROM links WHERE slug = $1;",
			slug,
		).Scan(
			&redirect_url, &click_count, &click_limit, &expires, &owner_id,
		)

		if err != nil {
			t.Fatalf("want: nil, got: %v\n", err)
		}

		if redirect_url != "https://google.com" {
			t.Fatalf("want: https://google.com, got: %s\n", redirect_url)
		}

		if click_count != 0 {
			t.Fatalf("want: 0, got: %d\n", click_count)
		}

		if click_limit != nil {
			t.Fatalf("want: nil, got: %v\n", click_limit)
		}

		if expires != nil {
			t.Fatalf("want: nil, got: %v\n", expires)
		}

		if owner_id != user_id {
			t.Fatalf("want: %d, got: %d\n", user_id, owner_id)
		}
	})
}
