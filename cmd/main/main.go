package main

import (
	"bytes"
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

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rohitxdev/go-api-template/internal/env"
	"github.com/rohitxdev/go-api-template/internal/handlers"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/swaggo/echo-swagger/example/docs"
)

var Env = env.Values

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	return nil
}

func getLogOutput() *os.File {
	logOutput := new(os.File)
	if Env.IS_DEV {
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

func getHTMLErrorString(statusCode int, errorMessage string) string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("error.templ").ParseFiles("./views/error.templ")
	if err != nil {
		log.Println("Error: Could not parse error template view", err.Error())
		return fmt.Sprintf("HTTP Error %v - %v", statusCode, errorMessage)
	}
	err = tmpl.Execute(buf, map[string]interface{}{"StatusCode": statusCode, "ErrorMessage": errorMessage})
	if err != nil {
		return err.Error()
	}
	return buf.String()
}

func main() {
	logOutput := getLogOutput()
	defer logOutput.Close()

	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("./views/*.templ")),
	}
	e.Validator = &CustomValidator{validator: validator.New()}

	e.Static("/", "./public")

	e.Pre(middleware.Recover())

	e.Pre(middleware.Secure())

	e.Pre(middleware.CORSWithConfig(middleware.CORSConfig{AllowCredentials: true}))

	e.Pre(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout:      5 * time.Second,
		ErrorMessage: getHTMLErrorString(503, "Timeout error. Server took too long to respond."),
	}))

	e.Pre(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${host} >>#${id} ${method} ${uri} from ${remote_ip} took ${latency_human} - status ${status} - sent ${bytes_out} bytes ${error}\n",
		Output: logOutput,
	}))

	e.Pre(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      30,
			Burst:     60,
			ExpiresIn: 1 * time.Minute,
		})}))

	e.GET("/api/docs/*", echoSwagger.WrapHandler)
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Skipper: func(c echo.Context) bool {
		return strings.Contains(c.Path(), "/metrics")
	}}))
	e.Pre(echoprometheus.NewMiddleware("echo"))
	e.GET("/metrics", echoprometheus.NewHandler())

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
				{Key: "Env", Value: env.Values.APP_ENV},
				{Key: "Host", Value: host},
				{Key: "PID", Value: strconv.Itoa(os.Getpid())},
				{Key: "OS", Value: runtime.GOOS},
			},
		})
	})

	handlers.MountRoutesOn(e)

	go func() {
		if err := e.StartTLS(Env.HOST+":"+Env.PORT, Env.TLS_CERT_PATH, Env.TLS_KEY_PATH); err != nil {
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
