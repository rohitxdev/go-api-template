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
	HOST                   = env.HOST
	PORT                   = env.PORT
	JWT_SECRET             = env.JWT_SECRET
	TLS_CERT_PATH          = env.TLS_CERT_PATH
	TLS_KEY_PATH           = env.TLS_KEY_PATH
	APP_ENV                = env.APP_ENV
	DB_URL                 = env.DB_URL
	GOOGLE_CLIENT_ID       = env.GOOGLE_CLIENT_ID
	GOOGLE_CLIENT_SECRET   = env.GOOGLE_CLIENT_SECRET
	FACEBOOK_CLIENT_ID     = env.FACEBOOK_CLIENT_ID
	FACEBOOK_CLIENT_SECRET = env.FACEBOOK_CLIENT_SECRET
	REDIS_HOST             = env.REDIS_HOST
	REDIS_PORT             = env.REDIS_PORT
	REDIS_USERNAME         = env.REDIS_USERNAME
	REDIS_PASSWORD         = env.REDIS_PASSWORD
	HTTPS                  = env.HTTPS
	IS_DEV                 = env.IS_DEV
)

var env = struct {
	HOST                   string `validate:"required,ip"`
	PORT                   string `validate:"required,gte=0"`
	JWT_SECRET             string `validate:"required"`
	TLS_CERT_PATH          string `validate:"required,filepath"`
	TLS_KEY_PATH           string `validate:"required,filepath"`
	APP_ENV                string `validate:"required,oneof=development production"`
	DB_URL                 string `validate:"required"`
	GOOGLE_CLIENT_ID       string
	GOOGLE_CLIENT_SECRET   string
	FACEBOOK_CLIENT_ID     string
	FACEBOOK_CLIENT_SECRET string
	REDIS_HOST             string
	REDIS_PORT             string
	REDIS_USERNAME         string
	REDIS_PASSWORD         string
	HTTPS                  bool
	IS_DEV                 bool
}{
	APP_ENV:                os.Getenv("APP_ENV"),
	HOST:                   os.Getenv("HOST"),
	PORT:                   os.Getenv("PORT"),
	JWT_SECRET:             os.Getenv("JWT_SECRET"),
	TLS_CERT_PATH:          os.Getenv("TLS_CERT_PATH"),
	TLS_KEY_PATH:           os.Getenv("TLS_KEY_PATH"),
	DB_URL:                 os.Getenv("DB_URL"),
	GOOGLE_CLIENT_ID:       os.Getenv("GOOGLE_CLIENT_ID"),
	GOOGLE_CLIENT_SECRET:   os.Getenv("GOOGLE_CLIENT_SECRET"),
	FACEBOOK_CLIENT_ID:     os.Getenv("FACEBOOK_CLIENT_ID"),
	FACEBOOK_CLIENT_SECRET: os.Getenv("FACEBOOK_CLIENT_SECRET"),
	REDIS_HOST:             os.Getenv("REDIS_HOST"),
	REDIS_PORT:             os.Getenv("REDIS_PORT"),
	REDIS_USERNAME:         os.Getenv("REDIS_USERNAME"),
	REDIS_PASSWORD:         os.Getenv("REDIS_PASSWORD"),
	HTTPS:                  os.Getenv("HTTPS") == "true",
	IS_DEV:                 os.Getenv("APP_ENV") != "production",
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

func loadEnv() {
	log.Println("Loading environment variables...")

	err := godotenv.Load(".env")
	if err != nil {
		//For testing. ".env" doesn't work when using 'go test' command so we need to use "../../.env"
		godotenv.Load("../../.env")
	}

	validate := validator.New()
	if err := validate.Struct(env); err != nil {
		log.Fatalln(err)
	}

	log.Println("Loaded environment variables successfully ✅")

	if env.IS_DEV {
		printEnv(env)
	}
}

func init() {
	loadEnv()
}
