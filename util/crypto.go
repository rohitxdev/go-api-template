package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/rohitxdev/go-api-template/env"
)

var (
	ErrTokenExpired   = errors.New("token expired")
	ErrMalformedToken = errors.New("malformed token")
)

var AccessTokenExpiresIn, RefreshTokenExpiresIn = func() (time.Duration, time.Duration) {
	accessTokenExpiresIn, err := time.ParseDuration(env.ACCESS_TOKEN_EXPIRES_IN)
	if err != nil {
		panic("could not parse access token expiration duration: " + err.Error())
	}
	refreshTokenExpiresIn, err := time.ParseDuration(env.REFRESH_TOKEN_EXPIRES_IN)
	if err != nil {
		panic("could not parse refresh token expiration duration: " + err.Error())
	}
	return accessTokenExpiresIn, refreshTokenExpiresIn
}()

// Encrypts data using AES algorithm. The key should be 16, 24, or 32 for 128, 192, or 256 bit encryption respectively.
func EncryptAES(data []byte, key []byte) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic("could not create cipher block: " + err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("could not create GCM: " + err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		panic("could not create nonce: " + err.Error())
	}
	//Append cipher to nonce and return nonce + cipher
	return gcm.Seal(nonce, nonce, data, nil)
}

// Decrypts data using AES algorithm. The key should be same key that was used to encrypt the data.
func DecryptAES(encryptedData []byte, key []byte) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic("could not create cipher block: " + err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("could not create GCM: " + err.Error())
	}
	nonceSize := gcm.NonceSize()

	//Get nonce from encrypted data
	nonce, cipher := encryptedData[:nonceSize], encryptedData[nonceSize:]
	data, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		panic("could not decrypt: " + err.Error())
	}
	return data
}

func GenerateJWT(userId uint, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{Id: fmt.Sprintf("%v", userId), ExpiresAt: time.Now().Add(expiresIn).Unix()})
	tokenString, err := token.SignedString([]byte(env.JWT_SECRET))
	if err != nil {
		return "", fmt.Errorf("could not get signed token string: %w", err)
	}
	return tokenString, nil
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
				return 0, ErrTokenExpired
			case jwt.ValidationErrorMalformed:
				return 0, ErrMalformedToken
			}
		}
		return 0, err
	}
	userId, _ := strconv.Atoi(claims.Id)
	return uint(userId), nil
}

func GenerateAccessAndRefreshTokens(userId uint) (string, string) {
	accessToken, _ := GenerateJWT(userId, AccessTokenExpiresIn)
	refreshToken, _ := GenerateJWT(userId, RefreshTokenExpiresIn)
	return accessToken, refreshToken
}
