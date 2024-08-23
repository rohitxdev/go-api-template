package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"golang.org/x/oauth2"

	"github.com/rohitxdev/go-api-template/util"
)

var oAuth2State = ulid.Make().String()

func (h *Handler) GetOAuth2UserEmail(c echo.Context, config *oauth2.Config, userDataEndpoint string) (string, error) {
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

type OAuth2LogInRequest struct {
	Provider string `param:"provider" validate:"required,oneof=google github discord"`
}

func (h *Handler) OAuth2LogIn(c echo.Context) error {
	req := new(OAuth2LogInRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}

	switch req.Provider {
	case "google":
		return c.Redirect(http.StatusTemporaryRedirect, h.config.GOOGLE_OAUTH2_CONFIG.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
	case "github":
		return c.Redirect(http.StatusTemporaryRedirect, h.config.GITHUB_OAUTH2_CONFIG.AuthCodeURL(oAuth2State, oauth2.ApprovalForce))
	default:
		return c.String(http.StatusUnprocessableEntity, "invalid provider")
	}
}

type OAuth2CallbackRequest struct {
	Provider string `param:"provider" validate:"required,oneof=google github discord"`
}

func (h *Handler) OAuth2Callback(c echo.Context) error {
	req := new(OAuth2CallbackRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}

	var err error
	var email string

	switch req.Provider {
	case "google":
		email, err = h.GetOAuth2UserEmail(c, h.config.GOOGLE_OAUTH2_CONFIG, "https://www.googleapis.com/oauth2/v2/userinfo")
	case "github":
		email, err = h.GetOAuth2UserEmail(c, h.config.GITHUB_OAUTH2_CONFIG, "https://api.github.com/user")
	default:
		return c.String(http.StatusUnprocessableEntity, "invalid provider")
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	user, err := h.repo.GetUserByEmail(c.Request().Context(), util.SanitizeEmail(email))
	// if err != nil {
	// if err == h.repo.ErrUserNotFound {
	// 	tokens, err := SignUp(ctx, email, uuid.NewString())
	// 	if err != nil {
	// 		return nil, fmt.Errorf("could not sign up user: %s", err.Error())
	// 	}
	// 	return err
	// }
	// return  err
	// }
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(uint(user.Id), h.config.ACCESS_TOKEN_EXPIRES_IN, h.config.REFRESH_TOKEN_EXPIRES_IN, h.config.JWT_SECRET)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	c.SetCookie(util.CreateLogInCookie(refreshToken, h.config.REFRESH_TOKEN_EXPIRES_IN))
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken})
}
