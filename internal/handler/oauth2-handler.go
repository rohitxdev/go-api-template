package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/internal/env"
	"github.com/rohitxdev/go-api-template/internal/service"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

var oauth2StateString = fmt.Sprintf("%x", rand.Int())

var googleOAuth2Config = &oauth2.Config{
	ClientID:     env.GOOGLE_CLIENT_ID,
	ClientSecret: env.GOOGLE_CLIENT_SECRET,
	Endpoint:     google.Endpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/google",
	Scopes:       []string{"openid email", "openid profile"},
}

var facebookOAuth2Config = &oauth2.Config{
	ClientID:     env.FACEBOOK_CLIENT_ID,
	ClientSecret: env.FACEBOOK_CLIENT_SECRET,
	Endpoint:     facebook.Endpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/facebook",
	Scopes:       []string{"email", "public_profile"},
}

func GetOAuth2User[T GoogleUser | FacebookUser](c echo.Context, config *oauth2.Config, userDataEndpoint string) (*T, error) {
	state := c.FormValue("state")
	code := c.FormValue("code")
	if state != oauth2StateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := config.Exchange(c.Request().Context(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	client := config.Client(c.Request().Context(), token)
	res, err := client.Get(userDataEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer res.Body.Close()
	contents, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	user := new(T)
	err = json.Unmarshal(contents, user)
	fmt.Println(string(contents))
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling response: %s", err.Error())
	}
	return user, nil
}

type GoogleUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func GoogleLogIn(c echo.Context) error {
	googleOAuth2Config.Endpoint.AuthURL = "https://accounts.google.com/o/oauth2/auth?prompt=consent"
	return c.Redirect(http.StatusTemporaryRedirect, googleOAuth2Config.AuthCodeURL(oauth2StateString))
}

func GoogleCallback(c echo.Context) error {
	googleUser, err := GetOAuth2User[GoogleUser](c, googleOAuth2Config, "https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	accessToken, refreshToken, err := service.UpsertUser(c.Request().Context(), googleUser.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken, "refresh_token": refreshToken})
}

type PictureData struct {
	Weight       int    `json:"height"`
	Width        int    `json:"width"`
	IsSilhouette bool   `json:"is_silhouette"`
	URL          string `json:"url"`
}

type FacebookUser struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data PictureData `json:"data"`
	} `json:"picture"`
}

func FacebookLogIn(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, facebookOAuth2Config.AuthCodeURL(oauth2StateString))
}

func FacebookCallback(c echo.Context) error {
	facebookUser, err := GetOAuth2User[FacebookUser](c, facebookOAuth2Config, "https://graph.facebook.com/v18.0/me?fields=id,name,email,picture")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	accessToken, refreshToken, err := service.UpsertUser(c.Request().Context(), facebookUser.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken, "refresh_token": refreshToken})
}
