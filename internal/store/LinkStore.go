package store

func GetShortenedURL(longUrl string) (string, error) {
	conn := Connect()

	conn.QueryRow("SELECT * from redirect_map")

	return "", nil
}
