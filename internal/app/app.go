package app

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/fs1g17/Mini-URL-Shortener/internal/store"
	"github.com/fs1g17/Mini-URL-Shortener/internal/user_context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type App struct {
	signingSecret string
	LinkStore     LinkStoreI
	UserStore     UserStoreI
}

type LinkStoreI interface {
	CreateShortenedURL(longUrl string, user_id int, config store.ShortUrlConfig) (string, error)
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
	LongURL    string  `json:"longURL"`
	Expire     *string `json:"expire"`
	ClickLimit *int    `json:"click_limit"`
}

func (app *App) PostLink(c echo.Context) error {
	var params PostLinkParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	user := user_context.FromContext(c.Request().Context())

	var expire *time.Time
	if params.Expire != nil {
		layout := "2006-01-02 15:04:05.000000 -0700 MST"
		parsedTime, err := time.Parse(layout, *params.Expire)
		expire = &parsedTime
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "time must be a valid time string like: '2012-10-31 15:50:13.793654 +0000 UTC'"})
		}
	}

	shortenedUrl, err := app.LinkStore.CreateShortenedURL(params.LongURL, user.UserID, store.ShortUrlConfig{Expire: expire, ClickLimit: params.ClickLimit})
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

var slugRe = regexp.MustCompile("^[A-Za-z0-9-_]{6}$")

func (p *GetRedirectParams) validate() bool {
	// check whether the slug matches
	match := slugRe.MatchString(p.Slug)
	return match
}

func (app *App) GetRedirect(c echo.Context) error {
	var params GetRedirectParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	if !params.validate() {
		return c.JSON(http.StatusBadRequest, "incorrect slug - must be base64 string of lenght 6")
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
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "incorrect username or password"})
		}
		return c.JSON(http.StatusInternalServerError, "something went sideways")
	}

	// generate jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": params.Username,
		"user_id":  user_id,
	})
	tokenString, err := token.SignedString([]byte(app.signingSecret))

	return c.JSON(http.StatusOK, map[string]string{"token": tokenString})
}
