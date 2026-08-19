package store

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetShortenedURL(longUrl string) (string, error) {
	conn := Connect()

	hasher := sha256.New()
	hasher.Write([]byte(longUrl))

	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	slug := hashString[len(hashString)-7:]

	_, err := conn.Exec("INSERT INTO redirect_map (slug, redirect_url) VALUES ($1, $2);", slug, longUrl)
	if err != nil {
		return "", err
	}

	return slug, nil
}
