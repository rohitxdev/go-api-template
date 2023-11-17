package handler

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/rohitxdev/go-api-template/internal/repo"
	"github.com/rohitxdev/go-api-template/internal/service"
	"github.com/rohitxdev/go-api-template/internal/util"
)

var (
	ErrUserNotLoggedIn = errors.New("user is not logged in")
)

/*----------------------------------- Get Access Token Handler ----------------------------------- */

func GetAccessToken(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	userId, err := util.VerifyJWT(refreshToken.Value)
	if err != nil {
		c.SetCookie(util.GetLogOutCookie())
		return c.String(http.StatusUnauthorized, err.Error())
	}
	accessToken := util.GenerateJWT(userId, util.AccessTokenExpiresIn)
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken})
}

/*----------------------------------- Log Out Handler ----------------------------------- */

func LogOut(c echo.Context) error {
	_, err := c.Cookie("refresh_token")
	if err != nil {
		return c.String(http.StatusBadRequest, ErrUserNotLoggedIn.Error())
	}
	c.SetCookie(util.GetLogOutCookie())
	return c.String(http.StatusOK, "Logged out")
}

/*----------------------------------- Log In Handler ----------------------------------- */

type LogInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func LogIn(c echo.Context) error {
	req := new(LogInRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	accessToken, refreshToken, err := service.LogIn(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	c.SetCookie(util.GetLogInCookie(refreshToken))
	return c.JSON(http.StatusOK, echo.Map{"access_token": accessToken})
}

/*----------------------------------- Sign Up Handler ----------------------------------- */

type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8,max=512"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqcsfield=Password"`
}

func SignUp(c echo.Context) error {
	req := new(SignUpRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	accessToken, refreshToken, err := service.SignUp(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	c.SetCookie(util.GetLogInCookie(refreshToken))
	return c.JSON(http.StatusCreated, echo.Map{"access_token": accessToken})
}

func SendPasswordChangeEmail(c echo.Context) error {
	user, ok := c.Get("user").(*repo.User)
	if !ok {
		return c.String(http.StatusUnauthorized, ErrUserNotLoggedIn.Error())
	}
	token := util.GenerateJWT(user.Id, time.Minute*10)
	u, _ := url.Parse("/v1/auth/change-password")
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return c.Redirect(http.StatusTemporaryRedirect, u.String())
}

/*----------------------------------- Forgot Password Handler ----------------------------------- */

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func ForgotPassword(c echo.Context) error {
	req := new(ForgotPasswordRequest)
	if err := util.BindAndValidate(c, req); err != nil {
		return err
	}
	user, err := repo.UserRepo.GetByEmail(c.Request().Context(), req.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return c.String(http.StatusNotFound, repo.ErrUserNotFound.Error())
		}
		return c.String(http.StatusInternalServerError, echo.ErrInternalServerError.Error())
	}
	token := util.GenerateJWT(user.Id, time.Minute*10)
	u, _ := url.Parse("https://192.168.0.110:8443/v1/auth/reset-password")
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	go func() {
		service.SendPasswordResetLink(u.String(), user.Email)
	}()
	return c.String(http.StatusOK, "sent password reset link to email")
}

/*----------------------------------- Reset Password Handler ----------------------------------- */

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
		return c.String(http.StatusOK, "success")
	}
}
