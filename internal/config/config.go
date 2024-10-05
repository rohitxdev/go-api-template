package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// This is set at build time.
var BuildId string

type Server struct {
	GoogleOAuth2Config    *oauth2.Config `json:"googleOAuth2Config"`
	GitHubOAuth2Config    *oauth2.Config `json:"githubOAuth2Config"`
	BuildId               string         `json:"buildId" validate:"required"`
	Env                   string         `json:"env" validate:"required,oneof=development production"`
	Host                  string         `json:"host" validate:"required,ip"`
	Port                  string         `json:"port" validate:"required,gte=0"`
	SessionSecret         string         `json:"sessionSecret" validate:"required"`
	DatabaseUrl           string         `json:"databaseUrl" validate:"required"`
	SmtpHost              string         `json:"smtpHost" validate:"required"`
	SmtpUsername          string         `json:"smtpUsername" validate:"required"`
	SmtpPassword          string         `json:"smtpPassword" validate:"required"`
	S3BucketName          string         `json:"s3BucketName"`
	S3Endpoint            string         `json:"s3Endpoint"`
	S3DefaultRegion       string         `json:"s3DefaultRegion"`
	AwsAccessKeyId        string         `json:"awsAccessKeyId"`
	AwsAccessKeySecret    string         `json:"awsAccessKeySecret"`
	GoogleClientId        string         `json:"googleClientId"`
	GoogleClientSecret    string         `json:"googleClientSecret"`
	AllowedOrigins        []string       `json:"allowedOrigins" validate:"required"`
	AccessTokenExpiresIn  time.Duration  `json:"accessTokenExpiresIn" validate:"required"`
	RefreshTokenExpiresIn time.Duration  `json:"refreshTokenExpiresIn" validate:"required"`
	ShutdownTimeout       time.Duration  `json:"shutdownTimeout" validate:"required"`
	RateLimitPerMinute    int            `json:"rateLimitPerMinute" validate:"required"`
	SmtpPort              int            `json:"smtpPort" validate:"required"`
}

type Client struct {
	Env string `json:"env" validate:"required,oneof=development production"`
}

func Load(envFilePaths ...string) (*Server, error) {
	if err := godotenv.Load(envFilePaths...); err != nil {
		return nil, fmt.Errorf("could not load env file(s): %w", err)
	}

	accessTokenExpiresIn, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_EXPIRES_IN"))
	if err != nil {
		return nil, fmt.Errorf("could not parse access token expiration duration: %w", err)
	}

	refreshTokenExpiresIn, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_EXPIRES_IN"))
	if err != nil {
		return nil, fmt.Errorf("could not parse refresh token expiration duration: %w", err)
	}

	smtpPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("could not parse SMTP port: %w", err)
	}

	shutdownTimeout, err := time.ParseDuration(os.Getenv("SHUTDOWN_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("could not parse shutdown timeout: %w", err)
	}

	rateLimitPerMinute, err := strconv.ParseInt(os.Getenv("RATE_LIMIT_PER_MINUTE"), 10, 8)
	if err != nil {
		return nil, fmt.Errorf("could not parse rate limit: %w", err)
	}

	c := Server{
		BuildId:               BuildId,
		Env:                   os.Getenv("ENV"),
		Host:                  os.Getenv("HOST"),
		Port:                  os.Getenv("PORT"),
		AllowedOrigins:        strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		S3BucketName:          os.Getenv("S3_BUCKET_NAME"),
		S3DefaultRegion:       os.Getenv("S3_DEFAULT_REGION"),
		AccessTokenExpiresIn:  accessTokenExpiresIn,
		RefreshTokenExpiresIn: refreshTokenExpiresIn,
		ShutdownTimeout:       shutdownTimeout,
		RateLimitPerMinute:    int(rateLimitPerMinute),
		SmtpPort:              int(smtpPort),
	}

	secretsFile := os.Getenv("SECRETS_FILE")
	if secretsFile != "" {
		data, err := os.ReadFile(secretsFile)
		if err != nil {
			return nil, fmt.Errorf("could not read secrets file: %w", err)
		}
		if err = json.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("could not unmarshal secrets file: %w", err)
		}
	}

	if err := validator.New().Struct(c); err != nil {
		return nil, fmt.Errorf("could not validate config: %w", err)
	}

	c.GoogleOAuth2Config = &oauth2.Config{
		ClientID:     c.GoogleClientId,
		ClientSecret: c.GoogleClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/v1/auth/oauth2/callback/google", c.Host+":"+c.Port),
		Scopes:       []string{"openid email", "openid profile"},
	}

	return &c, nil
}
