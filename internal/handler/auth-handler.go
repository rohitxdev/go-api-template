package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/internal/service"
)

func GetAccessToken(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusUnauthorized, "refresh_token not present in cookie")
	}
	userId, err := service.VerifyJWT(refreshToken.Value)
	if err != nil {
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    "refreshToken",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/v1/auth/refresh_token",
			MaxAge:   -1,
		})
		return c.String(http.StatusUnauthorized, err.Error())
	}
	accessToken, err := service.GenerateJWT(userId, service.AccessTokenExpiration)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken})
}

func LogOut(c echo.Context) error {
	_, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusBadRequest, "refresh_token not present in cookie")
	}
	return c.String(http.StatusOK, "Logged out")
}

type LogInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func LogIn(c echo.Context) error {
	req := new(LogInRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}
	accessToken, refreshToken, err := service.LogIn(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth/refresh_token",
		MaxAge:   int(service.RefreshTokenExpiration) / 1000000000,
	})
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken})
}

type SignUpRequest struct {
	LogInRequest
	ConfirmPassword string `json:"confirm_password" validate:"required,eqcsfield=Password"`
}

func SignUp(c echo.Context) error {
	req := new(SignUpRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return c.String(http.StatusUnprocessableEntity, err.Error())
	}
	accessToken, refreshToken, err := service.SignUp(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth/refresh_token",
		MaxAge:   int(service.RefreshTokenExpiration) / 1000000000,
	})
	return c.JSON(http.StatusCreated, echo.Map{"access_token": accessToken})
}
