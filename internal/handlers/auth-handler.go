package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/internal/services"
	"github.com/rohitxdev/go-api-template/internal/types"
	"golang.org/x/crypto/bcrypt"
)

func LogIn(c echo.Context) error {
	req := new(types.LogInRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}
	services.LogIn(req.Email, req.Password)
	return c.String(http.StatusOK, "Logged in")
}

func LogOut(c echo.Context) error {
	_, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusBadRequest, "refresh_token not present in cookie")
	}
	return c.String(http.StatusOK, "Logged out")
}

func SignUp(c echo.Context) error {
	req := new(types.SignUpRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, echo.Map{"hash": string(passwordHash), "token": services.GenerateAccessToken("rohit")})
}

func RefreshAccessToken(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusUnauthorized, "refresh_token not present in cookie")
	}
	return c.String(http.StatusOK, refreshToken.Value)
}
