package handlers

import (
	"github.com/labstack/echo/v4"
)

func MountRoutesOn(e *echo.Echo) {
	v1 := e.Group("/v1")
	auth := v1.Group("/auth")

	auth.POST("/log-in", LogIn)
	auth.POST("/log-out", LogOut)
	auth.POST("/sign-up", SignUp)
	auth.POST("/refresh-token", RefreshAccessToken)
}
