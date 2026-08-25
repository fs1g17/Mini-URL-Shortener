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

	_, err := ls.conn.Exec(context.Background(), "INSERT INTO redirect_map (slug, redirect_url, user_id) VALUES ($1, $2, $3);", slug, longUrl, user_id)
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
	var redirect_url string
	err := ls.conn.QueryRow(context.Background(), "SELECT redirect_url FROM redirect_map WHERE slug = $1;", slug).Scan(&redirect_url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", NoRedirectUrl
		}
		return "", err
	}

	return redirect_url, nil
}
