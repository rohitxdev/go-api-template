package handler

import (
	"github.com/labstack/echo/v4"
)

func MountRoutesOn(e *echo.Echo) {
	//API version 1
	v1 := e.Group("/v1")

	//Auth routes
	auth := v1.Group("/auth")
	auth.POST("/log-in", LogIn)
	auth.POST("/log-out", LogOut)
	auth.POST("/sign-up", SignUp)
	auth.POST("/forgot-password", ForgotPassword)
	auth.GET("/reset-password", ResetPassword)
	auth.POST("/reset-password", ResetPassword)
	auth.GET("/access-token", GetAccessToken)

	//OAuth2 login routes
	oauth2 := auth.Group("/oauth2")
	oauth2.GET("/google", GoogleLogIn)
	oauth2.GET("/github", GithubLogIn)
	oauth2.GET("/discord", DiscordLogIn)

	//OAuth2 callback routes
	callback := oauth2.Group("/callback")
	callback.GET("/google", GoogleCallback)
	callback.GET("/github", GithubCallback)
	callback.GET("/discord", DiscordCallback)

	//User routes
	users := v1.Group("/users")
	users.GET("", GetAllUsers, Auth(RoleAdmin))
	users.GET("/me", GetMe, Auth(RoleUser))
}

type paginatedQuery struct {
	Page uint `query:"page"`
}

func newPaginatedQuery() *paginatedQuery {
	return &paginatedQuery{Page: 1}
}
