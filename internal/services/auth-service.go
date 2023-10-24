package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lib/pq"
	"github.com/rohitxdev/go-api-template/internal/env"
)

var (
	secretKey            = []byte(env.Values.JWT_ACCESS_TOKEN_SECRET)
	errUserAlreadyExists = errors.New("user already exists")
	errDefault           = errors.New("unknown error")
)

func GetAccessToken() string {
	return "Here is your access token"
}

func LogIn(username string, password string) error {
	fmt.Println("Logged in user", username, password)
	return nil
}

func SignUp(email string, password string, first_name string, last_name string) error {
	_, err := db.Exec("INSERT INTO users(email,password,first_name,last_name) VALUES($1,$2,$3,$4)", email, password, first_name, last_name)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			switch err.Code.Name() {
			case "unique_violation":
				return errUserAlreadyExists
			default:
				return errDefault
			}
		}
	}
	return err
}

func GenerateAccessToken(id string) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{Id: id, ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix()})
	token, _ := t.SignedString([]byte("lol"))

	return token
}

func VerifyAccessToken(token string) error {
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}
	return nil
}
