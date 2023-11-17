package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/rohitxdev/go-api-template/internal/env"
	"github.com/rohitxdev/go-api-template/internal/service"
)

var oAuth2StateString = fmt.Sprintf("%x", rand.Int())

func init() {
	googleOAuth2Config.Endpoint.AuthURL += "?prompt=consent"
	githubOAuth2Config.Endpoint.AuthURL += "?prompt=consent"
	discordOAuth2Config.Endpoint.AuthURL += "?prompt=consent"
}

func GetOAuth2User[T GoogleUser | GithubUser | DiscordUser](c echo.Context, config *oauth2.Config, userDataEndpoint string) (*T, error) {
	state := c.FormValue("state")
	code := c.FormValue("code")
	if state != oAuth2StateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := config.Exchange(c.Request().Context(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}
	client := config.Client(c.Request().Context(), token)
	res, err := client.Get(userDataEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer res.Body.Close()
	contents, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}
	user := new(T)
	err = json.Unmarshal(contents, user)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling response: %w", err)
	}
	return user, nil
}

/*----------------------------------- Google Login Handler ----------------------------------- */

var googleOAuth2Config = &oauth2.Config{
	ClientID:     env.GOOGLE_CLIENT_ID,
	ClientSecret: env.GOOGLE_CLIENT_SECRET,
	Endpoint:     google.Endpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/google",
	Scopes:       []string{"openid email", "openid profile"},
}

type GoogleUser struct {
	Id    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	Image string `json:"picture,omitempty"`
}

func GoogleLogIn(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, googleOAuth2Config.AuthCodeURL(oAuth2StateString))
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

/*----------------------------------- Github Login Handler ----------------------------------- */

var githubOAuth2Config = &oauth2.Config{
	ClientID:     env.GITHUB_CLIENT_ID,
	ClientSecret: env.GITHUB_CLIENT_SECRET,
	Endpoint:     github.Endpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/github",
	Scopes:       []string{"read:user", "user:email"},
}

type GithubUser struct {
	Id    uint   `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	Image string `json:"avatar_url,omitempty"`
}

func GithubLogIn(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, githubOAuth2Config.AuthCodeURL(oAuth2StateString))
}

func GithubCallback(c echo.Context) error {
	githubUser, err := GetOAuth2User[GithubUser](c, githubOAuth2Config, "https://api.github.com/user")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	accessToken, refreshToken, err := service.UpsertUser(c.Request().Context(), githubUser.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken, "refresh_token": refreshToken})
}

/*----------------------------------- Discord Login Handler ----------------------------------- */

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

var discordOAuth2Config = &oauth2.Config{
	ClientID:     env.DISCORD_CLIENT_ID,
	ClientSecret: env.DISCORD_CLIENT_SECRET,
	Endpoint:     discordEndpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/discord",
	Scopes:       []string{"identify", "email"},
}

type DiscordUser struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"global_name,omitempty"`
	Email string `json:"email,omitempty"`
}

func DiscordLogIn(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, discordOAuth2Config.AuthCodeURL(oAuth2StateString))
}

func DiscordCallback(c echo.Context) error {
	discordUser, err := GetOAuth2User[DiscordUser](c, discordOAuth2Config, "https://discord.com/api/users/@me")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	accessToken, refreshToken, err := service.UpsertUser(c.Request().Context(), discordUser.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken, "refresh_token": refreshToken})
}
