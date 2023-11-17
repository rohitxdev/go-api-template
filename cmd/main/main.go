package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
	"golang.org/x/time/rate"

	"github.com/rohitxdev/go-api-template/internal/env"
	"github.com/rohitxdev/go-api-template/internal/handler"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Validator struct {
	validator *validator.Validate
}

func (cv *Validator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	return nil
}

func main() {
	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob(env.PROJECT_ROOT + "/templates/**/*.tmpl")),
	}

	e.Validator = &Validator{validator: validator.New()}

	e.GET("/docs/*", echoSwagger.WrapHandler)

	e.Static("/", "./public")

	e.Pre(middleware.Recover())

	e.Pre(middleware.Secure())

	e.Pre(middleware.CORS())

	e.Pre(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 5 * time.Second,
	}))

	/* Logger */

	logOutput := new(os.File)
	defer logOutput.Close()

	if env.IS_DEV {
		logOutput = os.Stdout
	} else {
		file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		logOutput = file
	}

	e.Pre(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${host} >>#${id} ${method} ${uri} from ${remote_ip} took ${latency_human} - status ${status} - sent ${bytes_out} bytes ${error}\n",
		Output: logOutput,
	}))

	/* Rate Limit */

	rateLimit, err := strconv.ParseUint(env.RATE_LIMIT_PER_MINUTE, 10, 8)
	if err != nil {
		panic(err)
	}

	e.Pre(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(rateLimit),
			ExpiresIn: 1 * time.Minute,
		})}))

	/* Gzip */

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Skipper: func(c echo.Context) bool {
		return !strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") || strings.HasSuffix(c.Path(), "/metrics")
	}}))

	/* Prometheus */

	e.Use(echoPrometheus.MetricsMiddleware())

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	/* Home page */

	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.tmpl", echo.Map{
			"Data": []struct {
				Key   string
				Value string
			}{
				{Key: "Name", Value: "Go + Echo App"},
				{Key: "Env", Value: env.APP_ENV},
				{Key: "Host", Value: host},
				{Key: "PID", Value: strconv.Itoa(os.Getpid())},
				{Key: "OS", Value: runtime.GOOS},
			},
		})
	})

	/* Start server */

	handler.MountRoutesOn(e)

	address := env.HOST + ":" + env.PORT
	if env.HTTPS {
		err = e.StartTLS(address, env.PROJECT_ROOT+env.TLS_CERT_PATH, env.PROJECT_ROOT+env.TLS_KEY_PATH)
	} else {
		err = e.Start(address)
	}
	if err != nil {
		panic(err)
	}
}
