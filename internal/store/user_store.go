package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var UsernameTakenErr = errors.New("username is already taken")

type UserStore struct {
	conn *pgx.Conn
}

func NewUserStore(conn *pgx.Conn) *UserStore {
	return &UserStore{
		conn: conn,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (us *UserStore) CreateUser(username string, password string) (int, error) {
	password_hash, err := hashPassword(password)
	if err != nil {
		return 0, err
	}

	var user_id int
	err = us.conn.QueryRow(context.Background(), "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id;", username, password_hash).Scan(&user_id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				// it's a unique violation
				return 0, UsernameTakenErr
			}
		}
	}

	return user_id, nil
}
