package handler

import (
	"github.com/labstack/echo/v4"
)

func MountRoutesOn(e *echo.Echo) {
	v1 := e.Group("/v1")

	auth := v1.Group("/auth")
	auth.POST("/log-in", LogIn)
	auth.POST("/log-out", LogOut)
	auth.POST("/sign-up", SignUp)
	auth.GET("/access-token", GetAccessToken)

	oauth2 := auth.Group("/oauth2")
	oauth2.GET("/google", GoogleLogIn)
	oauth2.GET("/facebook", FacebookLogIn)

	callback := oauth2.Group("/callback")
	callback.GET("/google", GoogleCallback)
	callback.GET("/facebook", FacebookCallback)

	users := v1.Group("/users")
	users.GET("", GetAllUsers, Auth(RoleAdmin))
	users.GET("/me", GetMe, Auth(RoleUser))
}
