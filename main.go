package main

import (
	"log"
	"net/http"

	"github.com/fs1g17/Mini-URL-Shortener/internal/store"
	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.GET("/hello-world", GetHelloWorld)
	e.POST("/link", PostLink)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

func GetHelloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

type PostLinkParams struct {
	LongURL string `json:"longURL"`
}

func PostLink(c echo.Context) error {
	var params PostLinkParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	shortenedUrl, err := store.GetShortenedURL(params.LongURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "something went sideways")
	}

	return c.JSON(http.StatusCreated, struct{ link string }{link: shortenedUrl})
}
