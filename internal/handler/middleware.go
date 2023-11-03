package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/internal/repo"
	"github.com/rohitxdev/go-api-template/internal/service"
)

type role uint8

const (
	RoleUser role = iota + 1
	RoleStaff
	RoleAdmin
)

var roleMap = map[string]role{
	"user":  RoleUser,
	"staff": RoleStaff,
	"admin": RoleAdmin,
}

func Auth(role role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.String(http.StatusUnauthorized, "invalid authorization header")
			}
			accessToken := strings.Split(authHeader, " ")[1]
			if accessToken == "" {
				return c.String(http.StatusUnauthorized, "invalid bearer token")
			}
			userId, err := service.VerifyJWT(accessToken)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}
			user, err := repo.UserRepo.GetById(c.Request().Context(), userId)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}
			if roleMap[user.Role] < role {
				return c.String(http.StatusForbidden, echo.ErrForbidden.Error())
			}
			c.Set("user", user)
			return next(c)
		}
	}
}
