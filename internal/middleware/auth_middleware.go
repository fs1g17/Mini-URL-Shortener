package middleware

import (
	"net/http"

	"github.com/fs1g17/Mini-URL-Shortener/internal/user_context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthMW struct {
	signingSecret string
}

func NewAuthMW(signingSecret string) *AuthMW {
	return &AuthMW{
		signingSecret: signingSecret,
	}
}

func (a *AuthMW) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		//headerString := c.Request().Header.Get("Authorization")

		authCookie, err := c.Request().Cookie("auth")
		if err != nil {
			// then user failed to authenticate
			return c.JSON(http.StatusUnauthorized, struct {
				Message string `json:"message"`
			}{Message: "unauthorized"})
		}

		tokenString := authCookie.Value
		// headerArr := strings.Split(headerString, " ")
		// tokenString := headerArr[1]

		// get just the token

		// here we have to validate signature, and grab user_id, stick it in context
		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (any, error) {
				return []byte(a.signingSecret), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)

		if err != nil {
			// then user failed to authenticate
			return c.JSON(http.StatusUnauthorized, struct {
				Message string `json:"message"`
			}{Message: "unauthorized"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusInternalServerError, struct {
				Message string `json:"message"`
			}{Message: "something went wrong"})
		}

		user_id_f, ok := claims["user_id"].(float64)
		if !ok {
			return c.JSON(http.StatusInternalServerError, struct {
				Message string `json:"message"`
			}{Message: "something went wrong"})
		}

		user_id := int(user_id_f)

		username, ok := claims["username"].(string)
		if !ok {
			return c.JSON(http.StatusInternalServerError, struct {
				Message string `json:"message"`
			}{Message: "something went wrong"})
		}

		req := c.Request()
		newContext := user_context.NewContext(c.Request().Context(), &user_context.ContextValue{Username: username, UserID: user_id})
		c.SetRequest(req.WithContext(newContext))
		return next(c)
	}
}
