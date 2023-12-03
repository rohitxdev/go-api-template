package util

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// BindAndValidate binds path params, query params and the request body into provided type `i` and validates provided `i`. The default binder binds body based on Content-Type header. Validator must be registered using `Echo#Validator`.
func BindAndValidate(c echo.Context, i any) (err error) {
	if err = c.Bind(i); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	binder := echo.DefaultBinder{}
	if err = binder.BindHeaders(c, i); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err = c.Validate(i); err != nil {
		c.String(http.StatusUnprocessableEntity, err.Error())
		return
	}
	return
}

var logOutCookie = &http.Cookie{
	Name:     "refresh_token",
	Secure:   true,
	HttpOnly: true,
	SameSite: http.SameSiteStrictMode,
	Path:     "/v1/auth/refresh_token",
	MaxAge:   0,
}

func CreateLogOutCookie() *http.Cookie {
	return logOutCookie
}

func CreateLogInCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth/refresh_token",
		MaxAge:   int(RefreshTokenExpiresIn / time.Second),
	}
}
