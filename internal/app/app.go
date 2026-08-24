package app

import (
	"errors"
	"net/http"

	"github.com/fs1g17/Mini-URL-Shortener/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type App struct {
	signingSecret string
	LinkStore     LinkStoreI
	UserStore     UserStoreI
}

type LinkStoreI interface {
	GetShortenedURL(longUrl string) (string, error)
	GetRedirectURL(slug string) (string, error)
}

type UserStoreI interface {
	CreateUser(username string, password string) (int, error)
	SignIn(username string, password string) (int, error)
}

func NewApp(signingSecret string) *App {
	conn := store.Connect()
	linkStore := store.NewLinkStore(conn)
	userStore := store.NewUserStore(conn)
	return &App{
		signingSecret: signingSecret,
		LinkStore:     linkStore,
		UserStore:     userStore,
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

type CreateUserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (app *App) CreateUser(c echo.Context) error {
	var params CreateUserParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	user_id, err := app.UserStore.CreateUser(params.Username, params.Password)
	if err != nil {
		if errors.Is(err, store.UsernameTakenErr) {
			return c.JSON(http.StatusConflict, struct {
				Message string `json:"message"`
			}{Message: "username already taken"})
		}
		return c.JSON(http.StatusInternalServerError, "couldn't create user")
	}

	return c.JSON(http.StatusCreated, struct {
		UserId int `json:"user_id"`
	}{UserId: user_id})
}

type SignInParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (app *App) SignIn(c echo.Context) error {
	var params SignInParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	user_id, err := app.UserStore.SignIn(params.Username, params.Password)
	if err != nil {
		if errors.Is(err, store.IncorrectInfoErr) {
			return c.JSON(http.StatusUnauthorized, "incorrect username or password")
		}
		return c.JSON(http.StatusInternalServerError, "something went sideways")
	}

	// generate jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": params.Username,
		"user_id":  user_id,
	})
	tokenString, err := token.SignedString([]byte(app.signingSecret))

	c.Response().Header().Set("Authorization", "Bearer "+tokenString)
	return c.JSON(http.StatusOK, "signed in")
}
