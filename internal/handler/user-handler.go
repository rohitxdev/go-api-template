package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rohitxdev/go-api-template/internal/repo"
	"github.com/rohitxdev/go-api-template/internal/util"
)

/*----------------------------------- Get Me Handler ----------------------------------- */

func GetMe(c echo.Context) error {
	user, ok := c.Get("user").(*repo.User)
	if !ok {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	return c.JSON(http.StatusOK, user)
}

/*----------------------------------- Get All Users Handler ----------------------------------- */

func GetAllUsers(c echo.Context) error {
	req := newPaginatedQuery()
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	users, err := repo.UserRepo.GetAll(c.Request().Context(), uint(req.Page))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, users)
}
