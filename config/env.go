package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var (
	HOST                     = config.HOST
	PORT                     = config.PORT
	JWT_SECRET               = config.JWT_SECRET
	ACCESS_TOKEN_EXPIRES_IN  = config.ACCESS_TOKEN_EXPIRES_IN
	REFRESH_TOKEN_EXPIRES_IN = config.REFRESH_TOKEN_EXPIRES_IN
	RATE_LIMIT_PER_MINUTE    = config.RATE_LIMIT_PER_MINUTE
	TLS_CERT_PATH            = config.TLS_CERT_PATH
	TLS_KEY_PATH             = config.TLS_KEY_PATH
	APP_ENV                  = config.APP_ENV
	DB_URL                   = config.DB_URL
	MONGODB_URL              = config.MONGODB_URL
	REDIS_HOST               = config.REDIS_HOST
	REDIS_PORT               = config.REDIS_PORT
	REDIS_USERNAME           = config.REDIS_USERNAME
	REDIS_PASSWORD           = config.REDIS_PASSWORD
	SMTP_HOST                = config.SMTP_HOST
	SMTP_PORT                = config.SMTP_PORT
	SMTP_USERNAME            = config.SMTP_USERNAME
	SMTP_PASSWORD            = config.SMTP_PASSWORD
	S3_BUCKET_NAME           = config.S3_BUCKET_NAME
	S3_DEFAULT_REGION        = config.S3_DEFAULT_REGION
	AWS_ACCESS_KEY_ID        = config.AWS_ACCESS_KEY_ID
	AWS_ACCESS_KEY_SECRET    = config.AWS_ACCESS_KEY_SECRET
	S3_ENDPOINT              = config.S3_ENDPOINT
	DISCORD_CLIENT_ID        = config.DISCORD_CLIENT_ID
	DISCORD_CLIENT_SECRET    = config.DISCORD_CLIENT_SECRET
	GOOGLE_CLIENT_ID         = config.GOOGLE_CLIENT_ID
	GOOGLE_CLIENT_SECRET     = config.GOOGLE_CLIENT_SECRET
	GITHUB_CLIENT_ID         = config.GITHUB_CLIENT_ID
	GITHUB_CLIENT_SECRET     = config.GITHUB_CLIENT_SECRET
	STRIPE_API_KEY           = config.STRIPE_API_KEY
	AMQP_URL                 = config.AMQP_URL
	HTTPS                    = config.HTTPS
	IS_DEV                   = config.IS_DEV
)

type appConfig struct {
	HOST                     string `validate:"required,ip"`
	PORT                     string `validate:"required,gte=0"`
	JWT_SECRET               string `validate:"required"`
	ACCESS_TOKEN_EXPIRES_IN  string `validate:"required"`
	REFRESH_TOKEN_EXPIRES_IN string `validate:"required"`
	RATE_LIMIT_PER_MINUTE    string `validate:"required"`
	TLS_CERT_PATH            string `validate:"required,filepath"`
	TLS_KEY_PATH             string `validate:"required,filepath"`
	APP_ENV                  string `validate:"required,oneof=development production"`
	DB_URL                   string `validate:"required"`
	MONGODB_URL              string `validate:"required"`
	REDIS_HOST               string `validate:"required"`
	REDIS_PORT               string `validate:"required"`
	REDIS_USERNAME           string `validate:"required"`
	REDIS_PASSWORD           string `validate:"required"`
	SMTP_HOST                string `validate:"required"`
	SMTP_PORT                string `validate:"required"`
	SMTP_USERNAME            string `validate:"required"`
	SMTP_PASSWORD            string `validate:"required"`
	S3_BUCKET_NAME           string
	S3_ENDPOINT              string
	S3_DEFAULT_REGION        string
	AWS_ACCESS_KEY_ID        string
	AWS_ACCESS_KEY_SECRET    string
	DISCORD_CLIENT_ID        string
	DISCORD_CLIENT_SECRET    string
	GOOGLE_CLIENT_ID         string
	GOOGLE_CLIENT_SECRET     string
	GITHUB_CLIENT_ID         string
	GITHUB_CLIENT_SECRET     string
	STRIPE_API_KEY           string
	AMQP_URL                 string
	HTTPS                    bool
	IS_DEV                   bool
}

func PrintEnv() {
	x := reflect.ValueOf(config)
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES")
	for i := 0; i < x.NumField(); i++ {
		key := x.Type().Field(i).Name
		value := x.Field(i)
		s := fmt.Sprintf("| %-40.40v | %-40.40v |", key, value)
		fmt.Println(strings.Repeat("-", utf8.RuneCountInString(s)))
		fmt.Println(s)
		if i == x.NumField()-1 {
			fmt.Println(strings.Repeat("-", utf8.RuneCountInString(s)))
		}
	}
	fmt.Println()
}

func loadConfig() appConfig {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	c := appConfig{
		APP_ENV:                  os.Getenv("APP_ENV"),
		HOST:                     os.Getenv("HOST"),
		PORT:                     os.Getenv("PORT"),
		JWT_SECRET:               os.Getenv("JWT_SECRET"),
		ACCESS_TOKEN_EXPIRES_IN:  os.Getenv("ACCESS_TOKEN_EXPIRES_IN"),
		REFRESH_TOKEN_EXPIRES_IN: os.Getenv("REFRESH_TOKEN_EXPIRES_IN"),
		RATE_LIMIT_PER_MINUTE:    os.Getenv("RATE_LIMIT_PER_MINUTE"),
		TLS_CERT_PATH:            os.Getenv("TLS_CERT_PATH"),
		TLS_KEY_PATH:             os.Getenv("TLS_KEY_PATH"),
		DB_URL:                   os.Getenv("DB_URL"),
		MONGODB_URL:              os.Getenv("MONGODB_URL"),
		REDIS_HOST:               os.Getenv("REDIS_HOST"),
		REDIS_PORT:               os.Getenv("REDIS_PORT"),
		REDIS_USERNAME:           os.Getenv("REDIS_USERNAME"),
		REDIS_PASSWORD:           os.Getenv("REDIS_PASSWORD"),
		SMTP_HOST:                os.Getenv("SMTP_HOST"),
		SMTP_PORT:                os.Getenv("SMTP_PORT"),
		SMTP_USERNAME:            os.Getenv("SMTP_USERNAME"),
		SMTP_PASSWORD:            os.Getenv("SMTP_PASSWORD"),
		S3_BUCKET_NAME:           os.Getenv("S3_BUCKET_NAME"),
		S3_ENDPOINT:              os.Getenv("S3_ENDPOINT"),
		S3_DEFAULT_REGION:        os.Getenv("S3_DEFAULT_REGION"),
		AWS_ACCESS_KEY_ID:        os.Getenv("AWS_ACCESS_KEY_ID"),
		AWS_ACCESS_KEY_SECRET:    os.Getenv("AWS_ACCESS_KEY_SECRET"),
		DISCORD_CLIENT_ID:        os.Getenv("DISCORD_CLIENT_ID"),
		DISCORD_CLIENT_SECRET:    os.Getenv("DISCORD_CLIENT_SECRET"),
		GOOGLE_CLIENT_ID:         os.Getenv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET:     os.Getenv("GOOGLE_CLIENT_SECRET"),
		GITHUB_CLIENT_ID:         os.Getenv("GITHUB_CLIENT_ID"),
		GITHUB_CLIENT_SECRET:     os.Getenv("GITHUB_CLIENT_SECRET"),
		STRIPE_API_KEY:           os.Getenv("STRIPE_API_KEY"),
		AMQP_URL:                 os.Getenv("AMQP_URL"),
		HTTPS:                    os.Getenv("HTTPS") == "true",
		IS_DEV:                   os.Getenv("APP_ENV") != "production",
	}

	if err := validator.New().Struct(c); err != nil {
		panic("app config validation failed: " + err.Error())
	}

	return c
}

var config = loadConfig()
