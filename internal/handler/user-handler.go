package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/internal/repo"
)

func getCurrentUser(c echo.Context) *repo.User {
	if user, ok := c.Get("user").(*repo.User); ok {
		return user
	}
	return nil
}

func GetMe(c echo.Context) error {
	user := getCurrentUser(c)
	if user == nil {
		return c.String(http.StatusUnauthorized, echo.ErrUnauthorized.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func GetAllUsers(c echo.Context) error {
	users, err := repo.UserRepo.GetAll(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(200, users)
}
