package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/rohitxdev/go-api-template/internal/env"
)

var (
	ErrInvalidOTP = errors.New("invalid/expired OTP")
)

var Redis = connectToRedis()

func connectToRedis() *redis.Client {
	log.Println("Connecting to redis...")
	rds := redis.NewClient(&redis.Options{
		Addr:     env.REDIS_HOST + ":" + env.REDIS_PORT,
		Username: env.REDIS_USERNAME,
		Password: env.REDIS_PASSWORD})
	if err := rds.Ping(context.TODO()).Err(); err != nil {
		log.Fatalln("Error when pinging redis: ", err.Error())
	}
	log.Println("Connected to redis successfully ✅")
	return rds
}

func GetTokensFromOTP(ctx context.Context, otp uint, userId uint) (string, string, error) {
	res := Redis.Get(ctx, fmt.Sprintf("login.%v", userId))
	err := res.Err()
	if err != nil {
		return "", "", ErrInvalidOTP
	}
	return GenerateAccessAndRefreshTokens(userId)
}
