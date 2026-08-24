package main

import (
	"log"
	"net/http"

	"github.com/fs1g17/Mini-URL-Shortener/internal/app"
	"github.com/fs1g17/Mini-URL-Shortener/internal/middleware"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	app := app.NewApp("secret")
	authMw := middleware.NewAuthMW("secret")

	e.GET("/hello-world", GetHelloWorld)
	e.POST("/sign-up", app.CreateUser)
	e.POST("/sign-in", app.SignIn)
	e.GET("/:slug", app.GetRedirect)

	auth := e.Group("/user")
	auth.POST("/link", app.PostLink, authMw.Process)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

func GetHelloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
