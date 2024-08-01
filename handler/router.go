package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/go-playground/validator"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oklog/ulid/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
	"golang.org/x/time/rate"

	"github.com/rohitxdev/go-api-template/config"
	_ "github.com/rohitxdev/go-api-template/docs"
	"github.com/rohitxdev/go-api-template/embedded"
	"github.com/rohitxdev/go-api-template/repo"
	"github.com/rohitxdev/go-api-template/service"
)

func RegisterRoutes(e *echo.Echo) {
	//	@BasePath	/v1
	v1 := e.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/log-in", LogIn)
			auth.POST("/log-out", LogOut)
			auth.POST("/sign-up", SignUp)
			auth.POST("/forgot-password", ForgotPassword)
			auth.GET("/reset-password", ResetPassword)
			auth.POST("/reset-password", ResetPassword)
			auth.GET("/access-token", GetAccessToken)
			auth.DELETE("/delete-account", DeleteAccount, Auth(RoleUser))
			auth.GET("/oauth2/:provider", OAuth2LogIn)
			auth.GET("/oauth2/callback/:provider", OAuth2Callback)
		}

		files := v1.Group("/files")
		{
			files.GET("/:file_name", GetFile)
			files.GET("", GetFileList)
			files.POST("", PutFile)
		}
	}
}

type echoTemplate struct {
	templates *template.Template
}

func (t *echoTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type echoValidator struct {
	validator *validator.Validate
}

func (v *echoValidator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	return nil
}

var logger = service.Logger

func GetEcho() *echo.Echo {
	e := echo.New()

	if config.IS_DEV {
		config.PrintEnv()
	} else {
		e.HideBanner = true
	}

	e.Renderer = &echoTemplate{
		templates: template.Must(template.ParseFS(embedded.FS, "templates/**/*.tmpl")),
	}

	e.Validator = &echoValidator{validator: validator.New()}

	e.Pre(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "public",
		Filesystem: http.FS(embedded.FS),
		HTML5:      true,
	}))

	e.Pre(middleware.Recover())

	e.Pre(middleware.Secure())

	if config.IS_DEV {
		e.Pre(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"*"}}))
	}

	e.Pre(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 5 * time.Second, Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/debug/pprof")
		},
	}))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Pre(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return ulid.Make().String()
		},
	}))

	e.Pre(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogHost:         true,
		LogRemoteIP:     true,
		LogProtocol:     true,
		LogURI:          true,
		LogMethod:       true,
		LogStatus:       true,
		LogLatency:      true,
		LogResponseSize: true,
		LogUserAgent:    true,
		LogReferer:      true,
		LogRequestID:    true,
		LogError:        true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			var userId uint
			user, ok := c.Get("user").(*repo.User)
			if ok && (user != nil) {
				userId = user.Id
			}

			logger.Info().
				Ctx(c.Request().Context()).
				Time("timestamp", v.StartTime).
				Str("host", v.Host).
				Str("client_ip", v.RemoteIP).
				Str("protocol", v.Protocol).
				Str("uri", v.URI).
				Str("method", v.Method).
				Int("status", v.Status).
				Dur("latency_ms", v.Latency.Round(time.Millisecond)).
				Int64("res_size_bytes", v.ResponseSize).
				Str("user_agent", v.UserAgent).
				Str("referer", v.Referer).
				Str("req_id", v.RequestID).
				Uint("user_id", userId).
				Err(v.Error).
				Msg("request")

			if config.IS_DEV {
				fmt.Println("-------------------------------------------------------")
			}
			return nil
		},
	}))

	rateLimit, err := strconv.ParseUint(config.RATE_LIMIT_PER_MINUTE, 10, 8)
	if err != nil {
		panic("could not parse rate limit: " + err.Error())
	}

	if rateLimit > 0 {
		e.Pre(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(rateLimit),
				ExpiresIn: time.Minute,
			})}))
	}

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Skipper: func(c echo.Context) bool {
		return !strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") || strings.HasSuffix(c.Request().URL.Path, "/metrics") || strings.HasSuffix(c.Request().URL.Path, "/swagger")
	}}))

	e.Pre(middleware.Decompress())

	//Do not use e.Pre() method for prometheuse middleware as it will show inaccurate metrics
	e.Use(echoPrometheus.MetricsMiddleware())

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	pprof.Register(e)

	host, err := os.Hostname()
	if err != nil {
		panic("could not get host name: " + err.Error())
	}

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.tmpl", echo.Map{
			"Data": []struct {
				Key   string
				Value string
			}{
				{Key: "Name", Value: "Go + Echo App"},
				{Key: "Env", Value: config.APP_ENV},
				{Key: "Host", Value: host},
				{Key: "PID", Value: strconv.Itoa(os.Getpid())},
				{Key: "OS", Value: runtime.GOOS},
			},
		})
	})

	RegisterRoutes(e)

	return e
}
