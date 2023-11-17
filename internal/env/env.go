package env

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var (
	HOST                     = env.HOST
	PORT                     = env.PORT
	JWT_SECRET               = env.JWT_SECRET
	ACCESS_TOKEN_EXPIRES_IN  = env.ACCESS_TOKEN_EXPIRES_IN
	REFRESH_TOKEN_EXPIRES_IN = env.REFRESH_TOKEN_EXPIRES_IN
	RATE_LIMIT_PER_MINUTE    = env.RATE_LIMIT_PER_MINUTE
	TLS_CERT_PATH            = env.TLS_CERT_PATH
	TLS_KEY_PATH             = env.TLS_KEY_PATH
	PROJECT_ROOT             = env.PROJECT_ROOT
	APP_ENV                  = env.APP_ENV
	DB_URL                   = env.DB_URL
	REDIS_HOST               = env.REDIS_HOST
	REDIS_PORT               = env.REDIS_PORT
	REDIS_USERNAME           = env.REDIS_USERNAME
	REDIS_PASSWORD           = env.REDIS_PASSWORD
	SMTP_HOST                = env.SMTP_HOST
	SMTP_PORT                = env.SMTP_PORT
	SMTP_USERNAME            = env.SMTP_USERNAME
	SMTP_PASSWORD            = env.SMTP_PASSWORD
	S3_BUCKET_NAME           = env.S3_BUCKET_NAME
	S3_DEFAULT_REGION        = env.S3_DEFAULT_REGION
	AWS_ACCESS_KEY_ID        = env.AWS_ACCESS_KEY_ID
	AWS_ACCESS_KEY_SECRET    = env.AWS_ACCESS_KEY_SECRET
	S3_ENDPOINT              = env.S3_ENDPOINT
	DISCORD_CLIENT_ID        = env.DISCORD_CLIENT_ID
	DISCORD_CLIENT_SECRET    = env.DISCORD_CLIENT_SECRET
	GOOGLE_CLIENT_ID         = env.GOOGLE_CLIENT_ID
	GOOGLE_CLIENT_SECRET     = env.GOOGLE_CLIENT_SECRET
	GITHUB_CLIENT_ID         = env.GITHUB_CLIENT_ID
	GITHUB_CLIENT_SECRET     = env.GITHUB_CLIENT_SECRET
	HTTPS                    = env.HTTPS
	IS_DEV                   = env.IS_DEV
)

var env = struct {
	HOST                     string `validate:"required,ip"`
	PORT                     string `validate:"required,gte=0"`
	JWT_SECRET               string `validate:"required"`
	ACCESS_TOKEN_EXPIRES_IN  string `validate:"required"`
	REFRESH_TOKEN_EXPIRES_IN string `validate:"required"`
	RATE_LIMIT_PER_MINUTE    string `validate:"required"`
	TLS_CERT_PATH            string `validate:"required,filepath"`
	TLS_KEY_PATH             string `validate:"required,filepath"`
	PROJECT_ROOT             string `validate:"required"`
	APP_ENV                  string `validate:"required,oneof=development production"`
	DB_URL                   string `validate:"required"`
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
	HTTPS                    bool
	IS_DEV                   bool
}{
	APP_ENV:                  os.Getenv("APP_ENV"),
	HOST:                     os.Getenv("HOST"),
	PORT:                     os.Getenv("PORT"),
	JWT_SECRET:               os.Getenv("JWT_SECRET"),
	ACCESS_TOKEN_EXPIRES_IN:  os.Getenv("ACCESS_TOKEN_EXPIRES_IN"),
	REFRESH_TOKEN_EXPIRES_IN: os.Getenv("REFRESH_TOKEN_EXPIRES_IN"),
	RATE_LIMIT_PER_MINUTE:    os.Getenv("RATE_LIMIT_PER_MINUTE"),
	TLS_CERT_PATH:            os.Getenv("TLS_CERT_PATH"),
	TLS_KEY_PATH:             os.Getenv("TLS_KEY_PATH"),
	PROJECT_ROOT:             os.Getenv("PROJECT_ROOT"),
	DB_URL:                   os.Getenv("DB_URL"),
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
	HTTPS:                    os.Getenv("HTTPS") == "true",
	IS_DEV:                   os.Getenv("APP_ENV") != "production",
}

func printEnv(v any) {
	x := reflect.ValueOf(v)
	fmt.Println()
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

func init() {
	log.Println("Loading environment variables...")

	if err := godotenv.Load(env.PROJECT_ROOT + "/.env"); err != nil {
		panic(err)
	}

	if err := validator.New().Struct(env); err != nil {
		panic(err)
	}

	log.Println("Loaded environment variables successfully ✅")

	if env.IS_DEV {
		printEnv(env)
	}
}
