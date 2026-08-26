package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var CollisionErr = errors.New("slug already exists")
var NoRedirectUrl = errors.New("no")

type LinkStore struct {
	conn *pgx.Conn
}

func NewLinkStore(conn *pgx.Conn) *LinkStore {
	return &LinkStore{
		conn: conn,
	}
}

func generateSlug() string {
	b := make([]byte, 4)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (ls *LinkStore) CreateShortenedURL(longUrl string, user_id int) (string, error) {
	slug := generateSlug()

	_, err := ls.conn.Exec(context.Background(), "INSERT INTO links (slug, redirect_url, owner_id) VALUES ($1, $2, $3);", slug, longUrl, user_id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				// it's a unique violation
				return "", CollisionErr
			}
		}
		return "", err
	}

	return slug, nil
}

func (ls *LinkStore) GetRedirectURL(slug string) (string, error) {
	// now we want to ensure expiry - we do that with an UPDATE ... WHERE, the WHERE clause will do the filering
	var redirect_url string
	var link_id int
	err := ls.conn.QueryRow(context.Background(), "UPDATE links SET click_count = click_count + 1 WHERE slug = $1 AND (click_limit IS NULL OR click_count < click_limit) RETURNING redirect_url, id;", slug).Scan(&redirect_url, &link_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", NoRedirectUrl
		}
		return "", err
	}

	//TODO: add the click event here
	_, err = ls.conn.Exec(context.Background(), "INSERT INTO click_events (link_id) VALUES ($1)", link_id)
	if err != nil {
		return "", err
	}

	return redirect_url, nil
}
