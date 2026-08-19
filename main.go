package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.GET("/hello-world", hello)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
