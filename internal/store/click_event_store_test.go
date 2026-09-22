package store

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func createTestUserAndLink(t *testing.T, userStore *UserStore, linkStore *LinkStore) {
	user_id, err := userStore.CreateUser("theo", "password")
	if err != nil {
		t.Fatalf("expected no error, got: %v\n", err)
	}

	full_url := "https://google.com"

	_, err = linkStore.CreateShortenedURL(full_url, user_id, ShortUrlConfig{})
	if err != nil {
		t.Fatalf("want: nil, got: %v\n", err)
	}
}

func emulateClick(clickEventStore *ClickEventStore, link_id int, clicked_at time.Time) {
	clickEventStore.conn.Exec(context.Background(), `INSERT INTO click_events (link_id, clicked_at) VALUES ($1, $2);`, link_id, clicked_at)
	clickEventStore.conn.Exec(context.Background(), `UPDATE links SET click_count = click_count+1 WHERE id = $1`, link_id)
}

func createTestClickEvents(clickEventStore *ClickEventStore) {
	clicks := []time.Time{
		time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 12, 1, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 12, 23, 29, 29, 999000, time.UTC),
		time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 13, 1, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 14, 1, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 14, 2, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 14, 3, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 14, 4, 0, 0, 0, time.UTC),
	}

	for _, click := range clicks {
		emulateClick(clickEventStore, 1, click)
	}
}

func TestGetHitsPerDay(t *testing.T) {
	clickEventStore := getTestClickEventStore(t)
	userStore := getTestUserStore(t)
	linkStore := getTestLinkStore(t)

	createTestUserAndLink(t, userStore, linkStore)
	createTestClickEvents(clickEventStore)

	t.Run("get clicks for 1 day", func(t *testing.T) {
		tests := []struct {
			day       int
			wantCount int
		}{
			{
				day:       12,
				wantCount: 3,
			},
			{
				day:       13,
				wantCount: 2,
			},
			{
				day:       14,
				wantCount: 4,
			},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("date: %d", tt.day), func(t *testing.T) {
				from := time.Date(2009, 1, tt.day, 0, 0, 0, 0, time.UTC)
				to := time.Date(2009, 1, tt.day, 0, 0, 0, 0, time.UTC)
				clicks, err := clickEventStore.GetHitsPerDay(1, from, to)
				if err != nil {
					t.Fatalf("failed to get clicks %v\n", err)
				}

				if len(clicks) != 1 {
					t.Fatalf("incorrect length for returned clicks")
				}

				count, ok := clicks[from]
				if !ok {
					t.Fatalf("got wrong date")
				}

				if count != tt.wantCount {
					t.Fatalf("count. want %d, got %d", tt.wantCount, count)
				}
			})
		}
	})

	t.Run("get clicks for 2 days", func(t *testing.T) {
		from := time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC)
		to := time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC)
		clicks, err := clickEventStore.GetHitsPerDay(1, from, to)
		if err != nil {
			t.Fatalf("failed to get clicks %v\n", err)
		}

		expectedClicks := map[time.Time]int{
			from: 3,
			to:   2,
		}

		if len(clicks) != len(expectedClicks) {
			t.Fatalf("want len: %d, got len: %d\n", len(expectedClicks), len(clicks))
		}

		for expectedKey, expectedValue := range expectedClicks {
			actualValue, ok := clicks[expectedKey]
			if !ok {
				t.Fatalf("want missing key: %v\n", expectedKey)
			}
			if actualValue != expectedValue {
				t.Fatalf("want: %v, got: %v", expectedValue, actualValue)
			}
		}
	})
}
