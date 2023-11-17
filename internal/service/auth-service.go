package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/rohitxdev/go-api-template/internal/repo"
	"github.com/rohitxdev/go-api-template/internal/util"
)

var (
	ErrIncorrectPassword = errors.New("incorrect password")
)

func LogIn(ctx context.Context, email string, password string) (string, string, error) {
	user, err := repo.UserRepo.GetByEmail(ctx, util.SanitizeEmail(email))
	if err != nil {
		return "", "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrIncorrectPassword
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(uint(user.Id))
	return accessToken, refreshToken, nil
}

func SignUp(ctx context.Context, email string, password string) (string, string, error) {
	user := new(repo.UserCore)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", "", err
	}
	user.Email = util.SanitizeEmail(email)
	user.PasswordHash = string(passwordHash)
	userId, err := repo.UserRepo.Create(ctx, user)
	if err != nil {
		return "", "", err
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(uint(userId))
	return accessToken, refreshToken, nil
}

// UpsertUser creates user if user does not exist and returns access and refresh tokens.
func UpsertUser(ctx context.Context, email string) (string, string, error) {
	user, err := repo.UserRepo.GetByEmail(ctx, util.SanitizeEmail(email))
	if err != nil {
		if err == repo.ErrUserNotFound {
			accessToken, refreshToken, err := SignUp(ctx, email, uuid.NewString())
			if err != nil {
				return "", "", fmt.Errorf("could not sign up user: %s", err.Error())
			}
			return accessToken, refreshToken, nil
		}
		return "", "", err
	}
	accessToken, refreshToken := util.GenerateAccessAndRefreshTokens(uint(user.Id))
	return accessToken, refreshToken, nil
}
