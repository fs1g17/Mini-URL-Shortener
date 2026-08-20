package app

import (
	"errors"
	"net/http"

	"github.com/fs1g17/Mini-URL-Shortener/internal/store"
	"github.com/jackc/pgx"
	"github.com/labstack/echo/v4"
)

type App struct {
	Conn      *pgx.Conn
	LinkStore LinkStoreI
}

type LinkStoreI interface {
	GetShortenedURL(longUrl string) (string, error)
	GetRedirectURL(slug string) (string, error)
}

func NewApp() *App {
	conn := store.Connect()
	linkStore := store.NewLinkStore(conn)
	return &App{
		Conn:      conn,
		LinkStore: linkStore,
	}
}

type PostLinkParams struct {
	LongURL string `json:"longURL"`
}

func (app *App) PostLink(c echo.Context) error {
	var params PostLinkParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	shortenedUrl, err := app.LinkStore.GetShortenedURL(params.LongURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "something went sideways")
	}

	type Response struct {
		Link string `json:"link"`
	}

	var response Response = Response{Link: shortenedUrl}

	return c.JSON(http.StatusCreated, response)
}

type GetRedirectParams struct {
	Slug string `param:"slug"`
}

func (app *App) GetRedirect(c echo.Context) error {
	var params GetRedirectParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	redirectUrl, err := app.LinkStore.GetRedirectURL(params.Slug)
	if err != nil {
		if errors.Is(err, store.NoRedirectUrl) {
			return c.JSON(http.StatusNotFound, "not found")
		}
		return c.JSON(http.StatusInternalServerError, "something went sideways")
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}
