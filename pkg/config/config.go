package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GOOGLE_OAUTH2_CONFIG     *oauth2.Config
	GITHUB_OAUTH2_CONFIG     *oauth2.Config
	HOST                     string `validate:"required,ip"`
	PORT                     string `validate:"required,gte=0"`
	JWT_SECRET               string `validate:"required"`
	APP_ENV                  string `validate:"required,oneof=development production"`
	DB_URL                   string `validate:"required"`
	MONGODB_URL              string `validate:"required"`
	REDIS_HOST               string `validate:"required"`
	REDIS_PORT               string `validate:"required"`
	REDIS_USERNAME           string `validate:"required"`
	REDIS_PASSWORD           string `validate:"required"`
	SMTP_HOST                string `validate:"required"`
	SMTP_USERNAME            string `validate:"required"`
	SMTP_PASSWORD            string `validate:"required"`
	S3_BUCKET_NAME           string
	S3_ENDPOINT              string
	S3_DEFAULT_REGION        string
	AWS_ACCESS_KEY_ID        string
	AWS_ACCESS_KEY_SECRET    string
	GOOGLE_CLIENT_ID         string
	GOOGLE_CLIENT_SECRET     string
	GITHUB_CLIENT_ID         string
	GITHUB_CLIENT_SECRET     string
	STRIPE_API_KEY           string
	NATS_URL                 string
	ACCESS_TOKEN_EXPIRES_IN  time.Duration `validate:"required"`
	REFRESH_TOKEN_EXPIRES_IN time.Duration `validate:"required"`
	SHUTDOWN_TIMEOUT         time.Duration `validate:"required"`
	RATE_LIMIT_PER_MINUTE    int           `validate:"required"`
	SMTP_PORT                int           `validate:"required"`
	IS_DEV                   bool
}

func Print(c Config) {
	x := reflect.ValueOf(c)
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

func Load(envFilePath string) (*Config, error) {
	if err := godotenv.Load(envFilePath); err != nil {
		fmt.Println("warning: could not load config file: " + err.Error())
	}

	accessTokenExpiresIn, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_EXPIRES_IN"))
	if err != nil {
		return nil, errors.Join(errors.New("parse access token expiration duration"), err)
	}

	refreshTokenExpiresIn, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_EXPIRES_IN"))
	if err != nil {
		return nil, errors.Join(errors.New("parse refresh token expiration duration"), err)
	}

	smtpPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 16)
	if err != nil {
		return nil, errors.Join(errors.New("parse SMTP port"), err)
	}

	shutdownTimeout, err := time.ParseDuration(os.Getenv("SHUTDOWN_TIMEOUT"))
	if err != nil {
		return nil, errors.Join(errors.New("parse shutdown timeout"), err)
	}

	rateLimitPerMinute, err := strconv.ParseInt(os.Getenv("RATE_LIMIT_PER_MINUTE"), 10, 8)
	if err != nil {
		return nil, errors.Join(errors.New("parse rate limit"), err)
	}

	c := Config{
		APP_ENV:                  os.Getenv("APP_ENV"),
		HOST:                     os.Getenv("HOST"),
		PORT:                     os.Getenv("PORT"),
		JWT_SECRET:               os.Getenv("JWT_SECRET"),
		DB_URL:                   os.Getenv("DB_URL"),
		MONGODB_URL:              os.Getenv("MONGODB_URL"),
		REDIS_HOST:               os.Getenv("REDIS_HOST"),
		REDIS_PORT:               os.Getenv("REDIS_PORT"),
		REDIS_USERNAME:           os.Getenv("REDIS_USERNAME"),
		REDIS_PASSWORD:           os.Getenv("REDIS_PASSWORD"),
		SMTP_HOST:                os.Getenv("SMTP_HOST"),
		SMTP_USERNAME:            os.Getenv("SMTP_USERNAME"),
		SMTP_PASSWORD:            os.Getenv("SMTP_PASSWORD"),
		S3_BUCKET_NAME:           os.Getenv("S3_BUCKET_NAME"),
		S3_ENDPOINT:              os.Getenv("S3_ENDPOINT"),
		S3_DEFAULT_REGION:        os.Getenv("S3_DEFAULT_REGION"),
		AWS_ACCESS_KEY_ID:        os.Getenv("AWS_ACCESS_KEY_ID"),
		AWS_ACCESS_KEY_SECRET:    os.Getenv("AWS_ACCESS_KEY_SECRET"),
		GOOGLE_CLIENT_ID:         os.Getenv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET:     os.Getenv("GOOGLE_CLIENT_SECRET"),
		GITHUB_CLIENT_ID:         os.Getenv("GITHUB_CLIENT_ID"),
		GITHUB_CLIENT_SECRET:     os.Getenv("GITHUB_CLIENT_SECRET"),
		STRIPE_API_KEY:           os.Getenv("STRIPE_API_KEY"),
		NATS_URL:                 os.Getenv("NATS_URL"),
		ACCESS_TOKEN_EXPIRES_IN:  accessTokenExpiresIn,
		REFRESH_TOKEN_EXPIRES_IN: refreshTokenExpiresIn,
		SHUTDOWN_TIMEOUT:         shutdownTimeout,
		RATE_LIMIT_PER_MINUTE:    int(rateLimitPerMinute),

		SMTP_PORT: int(smtpPort),
		IS_DEV:    os.Getenv("APP_ENV") != "production",
	}

	if err := validator.New().Struct(c); err != nil {
		return nil, errors.Join(errors.New("validate config"), err)
	}

	if c.GOOGLE_CLIENT_ID != "" && c.GOOGLE_CLIENT_SECRET != "" {
		c.GOOGLE_OAUTH2_CONFIG = &oauth2.Config{
			ClientID:     c.GOOGLE_CLIENT_ID,
			ClientSecret: c.GOOGLE_CLIENT_SECRET,
			Endpoint:     google.Endpoint,
			RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/google",
			Scopes:       []string{"openid email", "openid profile"},
		}
	}

	if c.GITHUB_CLIENT_ID != "" && c.GITHUB_CLIENT_SECRET != "" {
		c.GITHUB_OAUTH2_CONFIG = &oauth2.Config{
			ClientID:     c.GITHUB_CLIENT_ID,
			ClientSecret: c.GITHUB_CLIENT_SECRET,
			Endpoint:     github.Endpoint,
			RedirectURL:  "https://localhost:8443/v1/auth/oauth2/callback/github",
			Scopes:       []string{"read:user", "user:email"},
		}
	}

	return &c, nil
}
