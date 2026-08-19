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
		return c.String(http.StatusBadRequest, "bad request")
	}

	store.GetShortenedURL(params.LongURL)

	return c.String(http.StatusOK, "poop scoop")
}
