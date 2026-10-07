package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type ClickEventStore struct {
	conn *pgx.Conn
}

func NewClickEventStore(conn *pgx.Conn) *ClickEventStore {
	return &ClickEventStore{
		conn: conn,
	}
}

func (cs *ClickEventStore) GetHitsPerDay(link_id int, from time.Time, to time.Time) (map[time.Time]int, error) {
	query := `
		SELECT g::date AS day, COALESCE(c.total_count, 0) as total_count
		FROM generate_series(
			$1 AT TIME ZONE 'UTC',
			$2 AT TIME ZONE 'UTC',
			'1 day'::interval
		) AS g
		LEFT JOIN (
			SELECT date(clicked_at AT TIME ZONE 'UTC') AS day, count(clicked_at) as total_count
			FROM click_events
			WHERE clicked_at >= $1 AND clicked_at < $2 + interval '1 day'
			AND link_id = $3
			GROUP BY 1 
		) AS c
		ON g.g::date = c.day
		ORDER BY day;
	`

	rows, err := cs.conn.Query(context.Background(), query, from, to, link_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	response := make(map[time.Time]int, 0)

	for rows.Next() {
		var date time.Time
		var total_count int
		err := rows.Scan(&date, &total_count)
		if err != nil {
			return nil, err
		}

		response[date] = total_count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w\n", err)
	}

	return response, nil
}

type LinkRankResult struct {
	Slug       string
	ClickCount int
	Rank       int
	Percentage float64
}

func (cs *ClickEventStore) GetLinkRanks(owner_id int, from time.Time, to time.Time) ([]LinkRankResult, error) {
	query := `
	SELECT
				links.slug as slug,
				COUNT(click_events.id),
				DENSE_RANK () OVER (
								ORDER BY COUNT(click_events.id) DESC
				),
				COALESCE(
					100.0 * COUNT(click_events.id) / NULLIF(SUM(COUNT(click_events.id)) OVER (), 0),
					0
				) AS percent
	FROM
				links
	LEFT JOIN click_events 
	ON links.id = click_events.link_id
	AND clicked_at >= $2 AND clicked_at < $3::timestamptz + interval '1 day'
	WHERE links.owner_id = $1
	GROUP BY links.id
	ORDER BY dense_rank, slug;
	`

	rows, err := cs.conn.Query(context.Background(), query, owner_id, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	response := make([]LinkRankResult, 0)

	for rows.Next() {
		var slug string
		var clickCount int
		var rank int
		var percentage float64

		err := rows.Scan(&slug, &clickCount, &rank, &percentage)
		if err != nil {
			return nil, err
		}

		response = append(
			response,
			LinkRankResult{
				Slug:       slug,
				ClickCount: clickCount,
				Rank:       rank,
				Percentage: percentage,
			})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w\n", err)
	}

	return response, nil
}
