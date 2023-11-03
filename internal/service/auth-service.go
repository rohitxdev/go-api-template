package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rohitxdev/go-api-template/internal/env"
	"github.com/rohitxdev/go-api-template/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTokenExpiration  = time.Minute * 10
	RefreshTokenExpiration = time.Hour * 24 * 7
)

var (
	errIncorrectPassword = errors.New("incorrect password")
	errTokenExpired      = errors.New("token expired")
	errMalformedToken    = errors.New("malformed token")
)

func LogIn(ctx context.Context, email string, password string) (string, string, error) {
	user, err := repo.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", errIncorrectPassword
	}
	return GenerateAccessAndRefreshTokens(uint(user.Id))
}

func SignUp(ctx context.Context, email string, password string) (string, string, error) {
	user := new(repo.UserCore)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", "", err
	}
	user.Email = email
	user.PasswordHash = string(passwordHash)
	userId, err := repo.UserRepo.Create(ctx, user)
	if err != nil {
		return "", "", err
	}
	return GenerateAccessAndRefreshTokens(userId)
}

func GenerateJWT(userId uint, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{Id: fmt.Sprintf("%v", userId), ExpiresAt: time.Now().Add(expiresIn).Unix()})
	return token.SignedString([]byte(env.JWT_SECRET))
}

func VerifyJWT(token string) (uint, error) {
	claims := new(jwt.StandardClaims)
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(env.JWT_SECRET), nil
	})
	if err != nil {
		if err, ok := err.(*jwt.ValidationError); ok {
			switch err.Errors {
			case jwt.ValidationErrorExpired:
				return 0, errTokenExpired
			case jwt.ValidationErrorMalformed:
				return 0, errMalformedToken
			}
		}
		return 0, err
	}
	userId, _ := strconv.Atoi(claims.Id)
	return uint(userId), nil
}

func GenerateAccessAndRefreshTokens(userId uint) (string, string, error) {
	accessToken, _ := GenerateJWT(userId, AccessTokenExpiration)
	refreshToken, _ := GenerateJWT(userId, RefreshTokenExpiration)
	return accessToken, refreshToken, nil
}

// Creates user if does not exist and returns access and refresh tokens.
func UpsertUser(ctx context.Context, email string) (string, string, error) {
	user, err := repo.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == repo.ErrUserNotFound {
			accessToken, refreshToken, err := SignUp(ctx, email, "")
			if err != nil {
				return "", "", fmt.Errorf("could not sign up user: %s", err.Error())
			}
			return accessToken, refreshToken, nil
		}
	}
	accessToken, refreshToken, err := GenerateAccessAndRefreshTokens(uint(user.Id))
	if err != nil {
		return "", "", fmt.Errorf("could not log in user: %s", err.Error())
	}
	return accessToken, refreshToken, nil
}
