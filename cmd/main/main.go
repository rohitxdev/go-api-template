package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rohitxdev/go-api-template/internal/env"
	"github.com/rohitxdev/go-api-template/internal/handler"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/swaggo/echo-swagger/example/docs"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	return nil
}

func getLogOutput() *os.File {
	logOutput := new(os.File)
	if env.IS_DEV {
		logOutput = os.Stdout
	} else {
		file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		logOutput = file
	}
	return logOutput
}

//	@title			SAAS App API
//	@version		1.0
//	@description	This is a sample server Petstore server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host		petstore.swagger.io
// @BasePath	/v2
func main() {
	logOutput := getLogOutput()
	defer logOutput.Close()

	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("./templates/*.templ")),
	}

	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/docs/*", echoSwagger.WrapHandler)

	e.Static("/", "./public")

	e.Pre(middleware.Recover())

	e.Pre(middleware.Secure())

	e.Pre(middleware.CORS())

	e.Pre(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 5 * time.Second,
	}))

	e.Pre(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${host} >>#${id} ${method} ${uri} from ${remote_ip} took ${latency_human} - status ${status} - sent ${bytes_out} bytes ${error}\n",
		Output: logOutput,
	}))

	// e.Pre(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
	// 	Skipper: middleware.DefaultSkipper,
	// 	Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
	// 		Rate:      30,
	// 		Burst:     60,
	// 		ExpiresIn: 1 * time.Minute,
	// 	})}))

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Skipper: func(c echo.Context) bool {
		return !strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") || strings.HasSuffix(c.Path(), "/metrics")
	}}))

	e.Use(echoPrometheus.MetricsMiddleware())

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	host, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.templ", echo.Map{
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

	handler.MountRoutesOn(e)

	go func() {
		var err error
		address := env.HOST + ":" + env.PORT
		if env.HTTPS {
			e.StartTLS(address, env.TLS_CERT_PATH, env.TLS_KEY_PATH)
		} else {
			e.Start(address)
		}
		if err != nil {
			log.Fatalln(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("\nShutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
