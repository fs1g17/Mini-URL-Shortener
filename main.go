package main

import (
	"log"
	"net/http"

	"github.com/fs1g17/Mini-URL-Shortener/internal/app"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	app := app.NewApp()

	e.GET("/hello-world", GetHelloWorld)
	e.POST("/link", app.PostLink)
	e.POST("/user", app.CreateUser)
	e.GET("/:slug", app.GetRedirect)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

func GetHelloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
