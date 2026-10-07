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
	link1Clicks := []time.Time{
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

	for _, click := range link1Clicks {
		emulateClick(clickEventStore, 1, click)
	}

	link2Clicks := []time.Time{
		time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 13, 1, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 14, 1, 0, 0, 0, time.UTC),
		time.Date(2009, 1, 14, 2, 0, 0, 0, time.UTC),
	}

	for _, click := range link2Clicks {
		emulateClick(clickEventStore, 2, click)
	}
}

func TestGetHitsPerDay(t *testing.T) {
	userStore := getTestUserStore(t)
	linkStore := getTestLinkStore(t)
	clickEventStore := getTestClickEventStore(t)

	createTestUserAndLink(t, userStore, linkStore)
	createTestClickEvents(clickEventStore)

	t.Run("get clicks for 1 day", func(t *testing.T) {
		tests := []struct {
			day       int
			wantCount int
			linkId    int
		}{
			{
				day:       11,
				wantCount: 0,
				linkId:    1,
			},
			{
				day:       12,
				wantCount: 3,
				linkId:    1,
			},
			{
				day:       13,
				wantCount: 2,
				linkId:    1,
			},
			{
				day:       14,
				wantCount: 4,
				linkId:    1,
			},
			{
				day:       12,
				wantCount: 1,
				linkId:    2,
			},
			{
				day:       13,
				wantCount: 2,
				linkId:    2,
			},
			{
				day:       14,
				wantCount: 3,
				linkId:    2,
			},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("link: %d date: %d", tt.linkId, tt.day), func(t *testing.T) {
				from := time.Date(2009, 1, tt.day, 0, 0, 0, 0, time.UTC)
				to := time.Date(2009, 1, tt.day, 0, 0, 0, 0, time.UTC)
				clicks, err := clickEventStore.GetHitsPerDay(tt.linkId, from, to)
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

	t.Run("get clicks over range with empty days", func(t *testing.T) {
		from := time.Date(2009, 1, 10, 0, 0, 0, 0, time.UTC)
		mid := time.Date(2009, 1, 11, 0, 0, 0, 0, time.UTC)
		to := time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC)
		clicks, err := clickEventStore.GetHitsPerDay(1, from, to)
		if err != nil {
			t.Fatalf("failed to get clicks %v\n", err)
		}

		expectedClicks := map[time.Time]int{
			from: 0,
			mid:  0,
			to:   3,
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

	t.Run("to date before from", func(t *testing.T) {
		from := time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC)
		to := time.Date(2009, 1, 11, 0, 0, 0, 0, time.UTC)
		clicks, err := clickEventStore.GetHitsPerDay(1, from, to)
		if err != nil {
			t.Fatalf("failed to get clicks %v\n", err)
		}

		if len(clicks) != 0 {
			t.Fatalf("want len: 0, got len: %d\n", len(clicks))
		}
	})
}

func createTestUserLinksAndClicks(t *testing.T, userStore *UserStore, linkStore *LinkStore, clickEventStore *ClickEventStore, clickEvents [][]time.Time) (int, map[string]int) {
	user_id, err := userStore.CreateUser("theo", "password")
	if err != nil {
		t.Fatalf("expected no error, got: %v\n", err)
	}

	link_id_map := make(map[int]int, len(clickEvents))
	link_slug_index_map := make(map[string]int, len(clickEvents))

	for i := 0; i < len(clickEvents); i++ {
		slug, err := linkStore.CreateShortenedURL(fmt.Sprintf("https://example%d.com", i), user_id, ShortUrlConfig{})
		if err != nil {
			// do something
		}

		var link_id int
		err = linkStore.conn.QueryRow(context.Background(), "SELECT id FROM links WHERE slug = $1;", slug).Scan(&link_id)
		if err != nil {
			// do something
		}

		link_id_map[i] = link_id
		link_slug_index_map[slug] = i
	}

	for link_index, link_click_timestamps := range clickEvents {
		link_id := link_id_map[link_index]
		for _, clicked_at := range link_click_timestamps {
			emulateClick(clickEventStore, link_id, clicked_at)
		}
	}

	return user_id, link_slug_index_map
}

func TestGetLinkRanks(t *testing.T) {
	userStore := getTestUserStore(t)
	linkStore := getTestLinkStore(t)
	clickEventStore := getTestClickEventStore(t)

	clickEvents := [][]time.Time{
		{
			// 12th - rank 1
			time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 12, 1, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 12, 2, 0, 0, 0, time.UTC),
			// 13th - rank 2
			time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 13, 1, 0, 0, 0, time.UTC),
			// 14th - rank 3
			time.Date(2009, 1, 14, 0, 0, 0, 0, time.UTC),
		},
		{
			// 12th - rank 2
			time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 12, 1, 0, 0, 0, time.UTC),
			// 13th - rank 1
			time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 13, 1, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 13, 2, 0, 0, 0, time.UTC),
			// 14th - rank 2
			time.Date(2009, 1, 14, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 14, 1, 0, 0, 0, time.UTC),
		},
		{
			// 12th - rank 3
			time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC),
			// 13th - rank 3
			time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC),
			// 14th - rank 1
			time.Date(2009, 1, 14, 0, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 14, 1, 0, 0, 0, time.UTC),
			time.Date(2009, 1, 14, 2, 0, 0, 0, time.UTC),
		},
	}

	user_id, link_slug_index_map := createTestUserLinksAndClicks(t, userStore, linkStore, clickEventStore, clickEvents)

	t.Run("test ranking by date", func(t *testing.T) {
		tests := []struct {
			name          string
			from          time.Time
			to            time.Time
			expectedRank  []int
			expectedCount []int
		}{
			{
				name:          "12th",
				from:          time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC),
				to:            time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC),
				expectedRank:  []int{0, 1, 2},
				expectedCount: []int{3, 2, 1},
			},
			{
				name:          "13th",
				from:          time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC),
				to:            time.Date(2009, 1, 13, 0, 0, 0, 0, time.UTC),
				expectedRank:  []int{1, 0, 2},
				expectedCount: []int{3, 2, 1},
			},
			{
				name:          "14th",
				from:          time.Date(2009, 1, 14, 0, 0, 0, 0, time.UTC),
				to:            time.Date(2009, 1, 14, 0, 0, 0, 0, time.UTC),
				expectedRank:  []int{2, 1, 0},
				expectedCount: []int{3, 2, 1},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				linkRank, err := clickEventStore.GetLinkRanks(user_id, tt.from, tt.to)
				if err != nil {
					t.Fatalf("didn't expect error: %v\n", err)
				}

				for i, rankedLink := range linkRank {
					index := link_slug_index_map[rankedLink.Slug]

					if index != tt.expectedRank[i] {
						t.Fatalf("want: %d got: %d\n", tt.expectedRank[i], index)
					}

					if rankedLink.ClickCount != tt.expectedCount[i] {
						t.Fatalf("want: %d got: %d\n", tt.expectedCount[i], rankedLink.ClickCount)
					}
				}
			})
		}
	})

	t.Run("test ranking for empty date", func(t *testing.T) {
		linkRank, err := clickEventStore.GetLinkRanks(user_id, time.Date(2009, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2009, 1, 15, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("didn't expect error: %v\n", err)
		}

		if len(linkRank) != 3 {
			t.Fatalf("want: %d got: %d\n", 3, len(linkRank))
		}

		for _, rankedLink := range linkRank {
			if rankedLink.ClickCount != 0 {
				t.Fatalf("want: %d got: %d\n", 0, rankedLink.ClickCount)
			}
			if rankedLink.Percentage != 0 {
				t.Fatalf("want: %f got: %f\n", 0.0, rankedLink.Percentage)
			}
			if rankedLink.Rank != 1 {
				t.Fatalf("want: %d got: %d\n", 1, rankedLink.Rank)
			}
		}
	})

	t.Run("test link ranks for other user", func(t *testing.T) {
		linkRank, err := clickEventStore.GetLinkRanks(user_id+1, time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC), time.Date(2009, 1, 12, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("didn't expect error: %v\n", err)
		}

		if len(linkRank) != 0 {
			t.Fatalf("want: %d got: %d\n", 0, len(linkRank))
		}
	})
}
