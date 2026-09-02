package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCreateShortenedUrl(t *testing.T) {
	tests := []struct {
		Name       string
		Expire     *time.Time
		ClickLimit *int
	}{
		{
			Name: "should create shortened url with no expiration for existing user",
		},
		{
			Name:       "should create shortened url with click expiration for existing user",
			ClickLimit: new(10),
		},
		{
			Name:   "should create shortened url with expiry for existing user",
			Expire: new(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			linkStore := getTestLinkStore(t)
			userStore := getTestUserStore(t)

			user_id, err := userStore.CreateUser("theo", "password")
			if err != nil {
				t.Fatalf("expected no error, got: %v\n", err)
			}

			slug, err := linkStore.CreateShortenedURL("https://google.com", user_id, ShortUrlConfig{ClickLimit: tt.ClickLimit, Expire: tt.Expire})
			if err != nil {
				t.Fatalf("expected no error, got: %v\n", err)
			}

			var redirect_url string
			var click_count int
			var click_limit *int
			var expires *time.Time
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

			if click_limit == nil && tt.ClickLimit != nil || click_limit != nil && tt.ClickLimit == nil {
				t.Fatalf("want: %v, got: %v\n", tt.ClickLimit, click_limit)
			}

			if click_limit != nil && tt.ClickLimit != nil && *click_limit != *tt.ClickLimit {
				t.Fatalf("want: %v, got: %v\n", tt.ClickLimit, click_limit)
			}

			if expires == nil && tt.Expire != nil {
				t.Fatalf("want: %v, got: %v\n", tt.Expire, expires)
			}

			if expires != nil && tt.Expire == nil {
				t.Fatalf("want: %v, got: %v\n", tt.Expire, expires)
			}

			if expires != nil && !expires.Truncate(time.Microsecond).Equal(tt.Expire.Truncate(time.Microsecond)) {
				t.Fatalf("want: %v, got: %v\n", tt.Expire, expires)
			}

			if owner_id != user_id {
				t.Fatalf("want: %d, got: %d\n", user_id, owner_id)
			}
		})
	}
}

func TestRedirectUrl(t *testing.T) {
	t.Run("should give redirect for unexpired link", func(t *testing.T) {
		linkStore := getTestLinkStore(t)
		userStore := getTestUserStore(t)

		user_id, err := userStore.CreateUser("theo", "password")
		if err != nil {
			t.Fatalf("expected no error, got: %v\n", err)
		}

		full_url := "https://google.com"

		slug, err := linkStore.CreateShortenedURL(full_url, user_id, ShortUrlConfig{})
		if err != nil {
			t.Fatalf("want: nil, got: %v\n", err)
		}

		redirect_url, err := linkStore.GetRedirectURL(slug)
		if err != nil {
			t.Fatalf("want nil, got: %v\n", err)
		}

		if redirect_url != full_url {
			t.Fatalf("want: %s, got: %s\n", full_url, redirect_url)
		}

		var click_count int
		var link_id int
		linkStore.conn.QueryRow(context.Background(), "SELECT id, click_count FROM links WHERE slug = $1;", slug).Scan(&link_id, &click_count)
		if click_count != 1 {
			t.Fatalf("want: 1, got: %d\n", click_count)
		}

		var click_event_count int
		linkStore.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM click_events WHERE link_id = $1;", link_id).Scan(&click_event_count)
		if click_event_count != 1 {
			t.Fatalf("want: 1, got: %d\n", click_event_count)
		}
	})

	t.Run("should give redirect for link with future expiry", func(t *testing.T) {
		linkStore := getTestLinkStore(t)
		userStore := getTestUserStore(t)

		user_id, err := userStore.CreateUser("theo", "password")
		if err != nil {
			t.Fatalf("expected no error, got: %v\n", err)
		}

		full_url := "https://google.com"

		hourFromNow := time.Now().Add(1 * time.Hour)
		slug, err := linkStore.CreateShortenedURL(full_url, user_id, ShortUrlConfig{Expire: &hourFromNow})
		if err != nil {
			t.Fatalf("want: nil, got: %v\n", err)
		}

		redirect_url, err := linkStore.GetRedirectURL(slug)
		if err != nil {
			t.Fatalf("want nil, got: %v\n", err)
		}

		if redirect_url != full_url {
			t.Fatalf("want: %s, got: %s\n", full_url, redirect_url)
		}

		var click_count int
		var link_id int
		linkStore.conn.QueryRow(context.Background(), "SELECT id, click_count FROM links WHERE slug = $1;", slug).Scan(&link_id, &click_count)
		if click_count != 1 {
			t.Fatalf("want: 1, got: %d\n", click_count)
		}

		var click_event_count int
		linkStore.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM click_events WHERE link_id = $1;", link_id).Scan(&click_event_count)
		if click_event_count != 1 {
			t.Fatalf("want: 1, got: %d\n", click_event_count)
		}
	})

	t.Run("should not give redirect for click count expired link", func(t *testing.T) {
		linkStore := getTestLinkStore(t)
		userStore := getTestUserStore(t)

		user_id, err := userStore.CreateUser("theo", "password")
		if err != nil {
			t.Fatalf("expected no error, got: %v\n", err)
		}

		full_url := "https://google.com"

		slug, err := linkStore.CreateShortenedURL(full_url, user_id, ShortUrlConfig{ClickLimit: new(10)})
		if err != nil {
			t.Fatalf("want: nil, got: %v\n", err)
		}

		linkStore.conn.Exec(context.Background(), "UPDATE links SET click_count = 10 WHERE slug = $1;", slug)

		redirect_url, err := linkStore.GetRedirectURL(slug)
		if !errors.Is(err, NoRedirectUrl) {
			t.Fatalf("want: %v, got: %v\n", NoRedirectUrl, err)
		}

		if redirect_url != "" {
			t.Fatalf("want: %s, got: %s\n", "", redirect_url)
		}

		var click_count int
		var link_id int
		linkStore.conn.QueryRow(context.Background(), "SELECT id, click_count FROM links WHERE slug = $1;", slug).Scan(&link_id, &click_count)
		if click_count != 10 {
			t.Fatalf("want: 10, got: %d\n", click_count)
		}

		var click_event_count int
		linkStore.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM click_events WHERE link_id = $1;", link_id).Scan(&click_event_count)
		if click_event_count != 0 {
			t.Fatalf("want: 0, got: %d\n", click_event_count)
		}
	})

	t.Run("should not give redirect for expired link", func(t *testing.T) {
		linkStore := getTestLinkStore(t)
		userStore := getTestUserStore(t)

		user_id, err := userStore.CreateUser("theo", "password")
		if err != nil {
			t.Fatalf("expected no error, got: %v\n", err)
		}

		full_url := "https://google.com"

		slug, err := linkStore.CreateShortenedURL(full_url, user_id, ShortUrlConfig{Expire: new(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))})
		if err != nil {
			t.Fatalf("want: nil, got: %v\n", err)
		}

		redirect_url, err := linkStore.GetRedirectURL(slug)
		if !errors.Is(err, NoRedirectUrl) {
			t.Fatalf("want: %v, got: %v\n", NoRedirectUrl, err)
		}

		if redirect_url != "" {
			t.Fatalf("want: %s, got: %s\n", "", redirect_url)
		}

		var click_count int
		var link_id int
		linkStore.conn.QueryRow(context.Background(), "SELECT id, click_count FROM links WHERE slug = $1;", slug).Scan(&link_id, &click_count)
		if click_count != 0 {
			t.Fatalf("want: 0, got: %d\n", click_count)
		}

		var click_event_count int
		linkStore.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM click_events WHERE link_id = $1;", link_id).Scan(&click_event_count)
		if click_event_count != 0 {
			t.Fatalf("want: 0, got: %d\n", click_event_count)
		}
	})
}
