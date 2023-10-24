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
	Values = loadEnv()
)

type env struct {
	HOST                     string `validate:"required,ip"`
	PORT                     string `validate:"required,gte=0"`
	JWT_ACCESS_TOKEN_SECRET  string `validate:"required"`
	JWT_REFRESH_TOKEN_SECRET string `validate:"required"`
	TLS_CERT_PATH            string `validate:"required,filepath"`
	TLS_KEY_PATH             string `validate:"required,filepath"`
	APP_ENV                  string `validate:"required,oneof=development production"`
	DB_URL                   string `validate:"required"`
	IS_DEV                   bool
}

func printEnvVars(v env) {
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

func loadEnv() env {
	log.Println("Loading environment variables...")
	err := godotenv.Load(".env")
	if err != nil {
		//For testing. ".env" doesn't work when using 'go test' command so we need to use "../../.env"
		godotenv.Load("../../.env")
	}

	envVars := env{
		APP_ENV:                  os.Getenv("APP_ENV"),
		HOST:                     os.Getenv("HOST"),
		PORT:                     os.Getenv("PORT"),
		JWT_ACCESS_TOKEN_SECRET:  os.Getenv("JWT_ACCESS_TOKEN_SECRET"),
		JWT_REFRESH_TOKEN_SECRET: os.Getenv("JWT_REFRESH_TOKEN_SECRET"),
		TLS_CERT_PATH:            os.Getenv("TLS_CERT_PATH"),
		TLS_KEY_PATH:             os.Getenv("TLS_KEY_PATH"),
		DB_URL:                   os.Getenv("DB_URL"),
		IS_DEV:                   os.Getenv("APP_ENV") != "production",
	}

	validate := validator.New()
	if err := validate.Struct(envVars); err != nil {
		log.Fatalln(err)
	}

	log.Println("Loaded environment variables successfully ✅")
	if envVars.IS_DEV {
		printEnvVars(envVars)
	}

	return envVars
}
