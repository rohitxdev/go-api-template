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

	"github.com/rohitxdev/go-api-template/config"
	"github.com/rohitxdev/go-api-template/service"
	"github.com/rohitxdev/go-api-template/util"
)

var oAuth2State = ulid.Make().String()

func GetOAuth2UserEmail(c echo.Context, config *oauth2.Config, userDataEndpoint string) (string, error) {
	code := c.FormValue("code")
	state := c.FormValue("state")
	if state != oAuth2State {
		return "", fmt.Errorf("invalid oauth2 state")
	}
	token, err := config.Exchange(c.Request().Context(), code)
	if err != nil {
		return "", fmt.Errorf("code exchange failed: %w", err)
	}
	client := config.Client(c.Request().Context(), token)
	res, err := client.Get(userDataEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed getting user info: %w", err)
	}
	defer res.Body.Close()
	contents, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading response body: %w", err)
	}
	user := make(map[string]any)
	err = json.Unmarshal(contents, &user)
	if err != nil {
		return "", fmt.Errorf("failed unmarshalling response: %w", err)
	}
	email, ok := user["email"].(string)
	if email == "" || !ok {
		return "", fmt.Errorf("email is empty")
	}
	return email, nil
}

var googleOAuth2Config = &oauth2.Config{
	ClientID:     config.GOOGLE_CLIENT_ID,
	ClientSecret: config.GOOGLE_CLIENT_SECRET,
	Endpoint:     google.Endpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/google",
	Scopes:       []string{"openid email", "openid profile"},
}

var githubOAuth2Config = &oauth2.Config{
	ClientID:     config.GITHUB_CLIENT_ID,
	ClientSecret: config.GITHUB_CLIENT_SECRET,
	Endpoint:     github.Endpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/github",
	Scopes:       []string{"read:user", "user:email"},
}

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

var discordOAuth2Config = &oauth2.Config{
	ClientID:     config.DISCORD_CLIENT_ID,
	ClientSecret: config.DISCORD_CLIENT_SECRET,
	Endpoint:     discordEndpoint,
	RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/discord",
	Scopes:       []string{"identify", "email"},
}

type OAuth2LogInRequest struct {
	Provider string `param:"provider" validate:"required,oneof=google github discord"`
}

// @Summary		OAuth2 Login
// @Description	Log in with OAuth2
// @Tags			auth
// @Accept			application/json
// @Produce		application/json
// @Router			/auth/oauth2/{provider} [GET]
// @Param provider   path string true "OAuth2 provider"
// @Success		200	object	LogInResponse
// @Failure		500	string	any
func OAuth2LogIn(c echo.Context) error {
	req := new(OAuth2LogInRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}

	switch req.Provider {
	case "google":
		return c.Redirect(http.StatusTemporaryRedirect, googleOAuth2Config.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
	case "github":
		return c.Redirect(http.StatusTemporaryRedirect, githubOAuth2Config.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
	case "discord":
		return c.Redirect(http.StatusTemporaryRedirect, discordOAuth2Config.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
	default:
		return c.String(http.StatusUnprocessableEntity, "invalid provider")
	}
}

type OAuth2CallbackRequest struct {
	Provider string `param:"provider" validate:"required,oneof=google github discord"`
}

func OAuth2Callback(c echo.Context) error {
	req := new(OAuth2CallbackRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}

	var err error
	var email string

	switch req.Provider {
	case "google":
		email, err = GetOAuth2UserEmail(c, googleOAuth2Config, "https://www.googleapis.com/oauth2/v2/userinfo")
	case "github":
		email, err = GetOAuth2UserEmail(c, githubOAuth2Config, "https://api.github.com/user")
	case "discord":
		email, err = GetOAuth2UserEmail(c, discordOAuth2Config, "https://discord.com/api/users/@me")
	default:
		return c.String(http.StatusUnprocessableEntity, "invalid provider")
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	tokens, err := service.UpsertUser(c.Request().Context(), email)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	c.SetCookie(util.CreateLogInCookie(tokens.RefreshToken))
	return c.JSON(http.StatusOK, echo.Map{"access_token": tokens.AccessToken})
}
