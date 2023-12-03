package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rohitxdev/go-api-template/repo"
	"github.com/rohitxdev/go-api-template/util"
)

/*----------------------------------- Get Me Handler ----------------------------------- */

// @Summary		Get me
// @Description	Get info of current user
// @Tags			users
// @Param			Authorization	header	string	true	"Access token"
// @Accept			application/json
// @Produce		application/json
// @Router			/users [GET]
// @Success		200	{object}	repo.User
// @Failure		500	string		any
func GetMe(c echo.Context) error {
	user, ok := c.Get("user").(*repo.User)
	if !ok {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	return c.JSON(http.StatusOK, user)
}

// @Summary		Get all users
// @Description	Get all users
// @Tags			users
// @Param			Authorization	header	string	true	"Access token"
// @Accept			application/json
// @Produce		application/json
// @Router			/users/me [GET]
// @Success		200	{object}	[]repo.User
// @Failure		500	string		any
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
