package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/rohitxdev/go-api-template/repo"
	"github.com/rohitxdev/go-api-template/util"
)

var (
	ErrIncorrectPassword = errors.New("incorrect password")
)

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

/*----------------------------------- Log In ----------------------------------- */

func LogIn(ctx context.Context, email string, password string) (*tokens, error) {
	user, err := repo.UserRepo.GetByEmail(ctx, util.SanitizeEmail(email))
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrIncorrectPassword
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(uint(user.Id))
	return &tokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

/*----------------------------------- Sign Up ----------------------------------- */

func SignUp(ctx context.Context, email string, password string) (*tokens, error) {
	user := new(repo.UserCore)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}
	user.Email = util.SanitizeEmail(email)
	user.PasswordHash = string(passwordHash)
	userId, err := repo.UserRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(uint(userId))
	return &tokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

/*----------------------------------- Upsert User ----------------------------------- */

// UpsertUser creates user if user does not exist and returns access and refresh tokens.
func UpsertUser(ctx context.Context, email string) (*tokens, error) {
	user, err := repo.UserRepo.GetByEmail(ctx, util.SanitizeEmail(email))
	if err != nil {
		if err == repo.ErrUserNotFound {
			tokens, err := SignUp(ctx, email, uuid.NewString())
			if err != nil {
				return nil, fmt.Errorf("could not sign up user: %s", err.Error())
			}
			return tokens, nil
		}
		return nil, err
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(uint(user.Id))
	return &tokens{AccessToken: accessToken, RefreshToken: refreshToken}, err
}
