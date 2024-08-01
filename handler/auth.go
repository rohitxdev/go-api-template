package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/rohitxdev/go-api-template/config"
	"github.com/rohitxdev/go-api-template/repo"
	"github.com/rohitxdev/go-api-template/service"
	"github.com/rohitxdev/go-api-template/util"
)

var (
	ErrUserNotLoggedIn = errors.New("user is not logged in")
)

// @Summary		Get access token
// @Description	Get access token if user is logged in
// @Tags			auth
// @Accept			application/json
// @Produce		application/json
// @Router			/auth/access-token [GET]
// @Success		200	string	any
// @Failure		401	string	httputil.HTTPError
func GetAccessToken(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	userId, err := util.VerifyJWT(refreshToken.Value)
	if err != nil {
		c.SetCookie(util.CreateLogOutCookie())
		return c.String(http.StatusUnauthorized, err.Error())
	}
	accessToken, _ := util.GenerateJWT(userId, util.AccessTokenExpiresIn)
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken})
}

// @Summary		Log out
// @Description	Log out of application
// @Tags			auth
// @Produce		application/json
// @Router			/auth/log-out [POST]
// @Success		200
func LogOut(c echo.Context) error {
	_, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusBadRequest, ErrUserNotLoggedIn.Error())
	}
	c.SetCookie(util.CreateLogOutCookie())
	return c.String(http.StatusOK, "Logged out")
}

type LogInRequest struct {
	Email    string `form:"email" json:"email" validate:"required,email"`
	Password string `form:"password" json:"password" validate:"required"`
}

type LogInResponse struct {
	AccessToken string `json:"access_token"`
}

// @Summary		Log in
// @Description	Log into application
// @Tags			auth
// @Param			body	body	LogInRequest	true	"Body"
// @Accept			application/json
// @Produce		application/json
// @Router			/auth/log-in [POST]
// @Success		200	{object}	LogInResponse
// @Failure		401	string		httputil.HTTPError
func LogIn(c echo.Context) error {
	req := new(LogInRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	tokens, err := service.LogIn(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	c.SetCookie(util.CreateLogInCookie(tokens.RefreshToken))
	return c.JSON(http.StatusOK, LogInResponse{AccessToken: tokens.AccessToken})
}

type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8,max=512"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqcsfield=Password"`
}

type SignUpResponse struct {
	AccessToken string `json:"access_token"`
}

// @Summary		Sign up
// @Description	Sign up for application
// @Tags			auth
// @Accept			multipart/form-data
// @Produce		application/json
// @Router			/auth/sign-up [POST]
// @Success		200	{object}	SignUpResponse
func SignUp(c echo.Context) error {
	req := new(SignUpRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	tokens, err := service.SignUp(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	c.SetCookie(util.CreateLogInCookie(tokens.RefreshToken))
	return c.JSON(http.StatusCreated, SignUpResponse{AccessToken: tokens.AccessToken})
}

func SendPasswordChangeEmail(c echo.Context) error {
	user, ok := c.Get("user").(*repo.User)
	if !ok {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	token, _ := util.GenerateJWT(user.Id, time.Minute*10)
	u, _ := url.Parse("/v1/auth/change-password")
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return c.Redirect(http.StatusTemporaryRedirect, u.String())
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// @Summary		Send password reset email
// @Description	Send password reset email
// @Tags			auth
// @Param			body	body	ForgotPasswordRequest	true	"Body"
// @Accept			application/json
// @Produce		application/json
// @Router			/auth/forgot-password [POST]
// @Success		200	{object}	string
func ForgotPassword(c echo.Context) error {
	req := new(ForgotPasswordRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	user, err := repo.UserRepo.GetByEmail(c.Request().Context(), util.SanitizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return c.String(http.StatusNotFound, repo.ErrUserNotFound.Error())
		}
		return c.String(http.StatusInternalServerError, echo.ErrInternalServerError.Error())
	}
	token, _ := util.GenerateJWT(user.Id, time.Minute*10)
	u, _ := url.Parse(fmt.Sprintf("https://%s:%s/v1/auth/reset-password", config.HOST, config.PORT))
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	go func() {
		service.SendPasswordResetLink(u.String(), user.Email)
	}()
	return c.String(http.StatusOK, "sent password reset link to email")
}

type ResetPasswordRequest struct {
	Token       string `form:"token" query:"token" validate:"required"`
	NewPassword string `form:"new_password" validate:"required,min=8,max=512"`
}

func ResetPassword(c echo.Context) error {
	req := new(ResetPasswordRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	userId, err := util.VerifyJWT(req.Token)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	_, err = repo.UserRepo.GetById(c.Request().Context(), userId)
	if errors.Is(err, repo.ErrUserNotFound) {
		return c.String(http.StatusNotFound, repo.ErrUserNotFound.Error())
	}
	if c.Request().Method == "GET" {
		return c.Render(http.StatusOK, "change-password.tmpl", nil)
	} else {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
		err := repo.UserRepo.Update(c.Request().Context(), userId, map[string]any{
			"password_hash": string(hash),
		})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "reset password successfully")
	}
}

func DeleteAccount(c echo.Context) error {
	user, ok := c.Get("user").(*repo.User)
	if !ok {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	if err := repo.UserRepo.DeleteById(c.Request().Context(), user.Id); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "deleted account successfully")
}
