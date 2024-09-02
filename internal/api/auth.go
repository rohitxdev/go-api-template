package api

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rohitxdev/go-api-template/pkg/repo"
	"github.com/rohitxdev/go-api-template/pkg/util"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotLoggedIn = errors.New("user is not logged in")
)

func (h *Handler) GetAccessToken(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	userId, err := util.VerifyJWT(refreshToken.Value, h.config.JwtSecret)
	if err != nil {
		c.SetCookie(createLogOutCookie())
		return c.String(http.StatusUnauthorized, err.Error())
	}
	accessToken, _ := util.GenerateJWT(userId, h.config.AccessTokenExpiresIn, h.config.JwtSecret)
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken})
}

func (h *Handler) LogOut(c echo.Context) error {
	_, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusBadRequest, ErrUserNotLoggedIn.Error())
	}
	c.SetCookie(createLogOutCookie())
	return c.String(http.StatusOK, "Logged out")
}

type LogInRequest struct {
	Email    string `form:"email" json:"email" validate:"required,email"`
	Password string `form:"password" json:"password" validate:"required"`
}

type LogInResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) LogIn(c echo.Context) error {
	req := new(LogInRequest)
	if err := bindAndValidate(c, req); err != nil {
		return err
	}
	user, err := h.repo.GetUserByEmail(c.Request().Context(), util.SanitizeEmail(req.Email))
	if err != nil {
		return err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(user.Id, h.config.AccessTokenExpiresIn, h.config.RefreshTokenExpiresIn, h.config.JwtSecret)
	c.SetCookie(createLogInCookie(refreshToken, h.config.RefreshTokenExpiresIn))
	return c.JSON(http.StatusOK, LogInResponse{AccessToken: accessToken})
}

type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8,max=512"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqcsfield=Password"`
}

type SignUpResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) SignUp(c echo.Context) error {
	req := new(SignUpRequest)
	if err := bindAndValidate(c, req); err != nil {
		return err
	}
	user := new(repo.UserCore)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return err
	}
	user.Email = util.SanitizeEmail(req.Email)
	user.PasswordHash = string(passwordHash)
	userId, err := h.repo.CreateUser(c.Request().Context(), user)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(userId, h.config.AccessTokenExpiresIn, h.config.RefreshTokenExpiresIn, h.config.JwtSecret)
	c.SetCookie(createLogInCookie(refreshToken, h.config.RefreshTokenExpiresIn))
	return c.JSON(http.StatusCreated, SignUpResponse{AccessToken: accessToken})
}

func (h *Handler) SendPasswordChangeEmail(c echo.Context) error {
	user, ok := c.Get("user").(*repo.User)
	if !ok {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	token, _ := util.GenerateJWT(user.Id, time.Minute*10, h.config.JwtSecret)
	u, _ := url.Parse("/v1/auth/change-password")
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return c.Redirect(http.StatusTemporaryRedirect, u.String())
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `form:"token" query:"token" validate:"required"`
	NewPassword string `form:"new_password" validate:"required,min=8,max=512"`
}

func (h *Handler) ResetPassword(c echo.Context) error {
	req := new(ResetPasswordRequest)
	if err := bindAndValidate(c, req); err != nil {
		return err
	}
	userId, err := util.VerifyJWT(req.Token, h.config.JwtSecret)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	_, err = h.repo.GetUserById(c.Request().Context(), userId)
	if errors.Is(err, repo.ErrUserNotFound) {
		return c.String(http.StatusNotFound, repo.ErrUserNotFound.Error())
	}
	if c.Request().Method == "GET" {
		return c.Render(http.StatusOK, "change-password.tmpl", nil)
	} else {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
		err := h.repo.Update(c.Request().Context(), userId, map[string]any{
			"password_hash": string(hash),
		})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "reset password successfully")
	}
}

func (h *Handler) DeleteAccount(c echo.Context) error {
	user, ok := c.Get("user").(*repo.User)
	if !ok {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	if err := h.repo.DeleteUserById(c.Request().Context(), user.Id); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "deleted account successfully")
}
