package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/rohitxdev/go-api-template/internal/env"
)

var (
	ErrTokenExpired   = errors.New("token expired")
	ErrMalformedToken = errors.New("malformed token")
)

var (
	AccessTokenExpiresIn  time.Duration
	RefreshTokenExpiresIn time.Duration
)

func init() {
	var err error
	AccessTokenExpiresIn, err = time.ParseDuration(env.ACCESS_TOKEN_EXPIRES_IN)
	if err != nil {
		panic(err)
	}
	RefreshTokenExpiresIn, err = time.ParseDuration(env.REFRESH_TOKEN_EXPIRES_IN)
	if err != nil {
		panic(err)
	}
}

// Encrypts data using AES algorithm. The key should be 16, 24, or 32 for 128, 192, or 256 bit encryption respectively.
func EncryptAES(data []byte, key []byte) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Panic("could not create cipher block:", err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Panic("could not create GCM:", err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		log.Panic("could not create nonce:", err.Error())
	}
	//Append cipher to nonce and return nonce + cipher
	return gcm.Seal(nonce, nonce, data, nil)
}

// Decrypts data using AES algorithm. The key should be same key that was used to encrypt the data.
func DecryptAES(encryptedData []byte, key []byte) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Panic("could not create cipher block:", err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Panic("could not create GCM:", err.Error())
	}
	nonceSize := gcm.NonceSize()

	//Get nonce from encrypted data
	nonce, cipher := encryptedData[:nonceSize], encryptedData[nonceSize:]
	data, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		log.Panic("could not decrypt:", err.Error())
	}
	return data
}

func GenerateJWT(userId uint, expiresIn time.Duration) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{Id: fmt.Sprintf("%v", userId), ExpiresAt: time.Now().Add(expiresIn).Unix()})
	tokenString, err := token.SignedString([]byte(env.JWT_SECRET))
	if err != nil {
		panic(err)
	}
	return tokenString
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
	accessToken := GenerateJWT(userId, AccessTokenExpiresIn)
	refreshToken := GenerateJWT(userId, RefreshTokenExpiresIn)
	return accessToken, refreshToken
}
