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
		SELECT g::date AS date, COALESCE(c.total_count, 0) as total_count
		FROM generate_series(
			$1::timestamp,
			$2::timestamp,
			'1 day'::interval
		) AS g
		LEFT JOIN (
			SELECT date(clicked_at), count(clicked_at) as total_count
			FROM click_events
			WHERE clicked_at BETWEEN $1 AND $2
			GROUP BY date(clicked_at)
			ORDER BY date(clicked_at)
		) AS c
		ON g.g = c.date;
	`

	rows, err := cs.conn.Query(context.Background(), query, from, to)
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
