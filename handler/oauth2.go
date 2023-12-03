package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/rohitxdev/go-api-template/env"
	"github.com/rohitxdev/go-api-template/service"
	"github.com/rohitxdev/go-api-template/util"
)

var oAuth2State = ulid.Make().String()

func GetOAuth2User[T GoogleUser | GithubUser | DiscordUser](c echo.Context, config *oauth2.Config, userDataEndpoint string) (*T, error) {
	code := c.FormValue("code")
	state := c.FormValue("state")
	if state != oAuth2State {
		return nil, fmt.Errorf("invalid oauth2 state")
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

func OAuth2Callback(c echo.Context, email string) error {
	tokens, err := service.UpsertUser(c.Request().Context(), email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	c.SetCookie(util.CreateLogInCookie(tokens.RefreshToken))
	return c.JSON(http.StatusOK, echo.Map{"access_token": tokens.AccessToken})
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

// @Summary		Google Login
// @Description	Log in with Google
// @Tags			oauth2
// @Accept			application/json
// @Produce		application/json
// @Router			/auth/oauth2/google [GET]
// @Success		200	object	LogInResponse
// @Failure		500	string	any
func GoogleLogIn(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, googleOAuth2Config.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
}

func GoogleCallback(c echo.Context) error {
	googleUser, err := GetOAuth2User[GoogleUser](c, googleOAuth2Config, "https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return OAuth2Callback(c, googleUser.Email)

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

// @Summary		Github Login
// @Description	Log in with Github
// @Tags			oauth2
// @Accept			application/json
// @Produce		application/json
// @Router			/auth/oauth2/github [GET]
// @Success		200	object	LogInResponse
// @Failure		500	string	any
func GithubLogIn(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, githubOAuth2Config.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
}

func GithubCallback(c echo.Context) error {
	githubUser, err := GetOAuth2User[GithubUser](c, githubOAuth2Config, "https://api.github.com/user")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return OAuth2Callback(c, githubUser.Email)
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

// @Summary		Discord Login
// @Description	Log in with Discord
// @Tags			oauth2
// @Accept			application/json
// @Produce		application/json
// @Router			/auth/oauth2/discord [GET]
// @Success		200	object	LogInResponse
// @Failure		500	string	any
func DiscordLogIn(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, discordOAuth2Config.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
}

func DiscordCallback(c echo.Context) error {
	discordUser, err := GetOAuth2User[DiscordUser](c, discordOAuth2Config, "https://discord.com/api/users/@me")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return OAuth2Callback(c, discordUser.Email)
}
