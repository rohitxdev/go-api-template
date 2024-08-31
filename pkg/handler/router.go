package handler

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rohitxdev/go-api-template/pkg/id"
	"github.com/rohitxdev/go-api-template/pkg/logger"
	"github.com/rohitxdev/go-api-template/pkg/repo"
	"golang.org/x/time/rate"

	"github.com/goccy/go-json"
)

// Custom view renderer

type echoTemplate struct {
	templates *template.Template
}

func (t echoTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Custom request validator

type echoValidator struct {
	validator *validator.Validate
}

func (v echoValidator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	return nil
}

// Custom JSON serializer & deserializer

type echoJSONSerializer struct{}

func (s echoJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

func (s echoJSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	err := json.NewDecoder(c.Request().Body).Decode(i)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
	} else if se, ok := err.(*json.SyntaxError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}
	return err
}

func NewRouter(h *Handler) (*echo.Echo, error) {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.JSONSerializer = echoJSONSerializer{}

	e.Renderer = echoTemplate{
		templates: template.Must(template.ParseFS(h.staticFS, "templates/**/*.tmpl")),
	}

	e.Validator = echoValidator{
		validator: validator.New(),
	}

	e.Pre(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "public",
		Filesystem: http.FS(h.staticFS),
	}))

	e.Pre(middleware.Recover())

	e.Pre(middleware.Secure())

	if h.config.IS_DEV {
		e.Pre(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"*"}}))
	}

	e.Pre(middleware.CSRF())

	e.Pre(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 5 * time.Second, Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/debug/pprof")
		},
	}))

	e.Pre(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return id.New(id.Request)
		},
	}))

	// HTTP Logger

	reqLoggerOpts := logger.HandlerOpts{
		TimeFormat: time.RFC3339,
		Level:      slog.LevelInfo,
		NoColor:    !h.config.IS_DEV,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch {
			case a.Key == "user_id" && a.Value.Int64() == 0:
				fallthrough
			case a.Value.String() == "":
				fallthrough
			case a.Value.Equal(slog.AnyValue(nil)):
				return slog.Attr{}
			}
			return a
		},
	}

	if h.config.IS_DEV {
		reqLoggerOpts.TimeFormat = time.Kitchen
	}

	logger := slog.New(logger.NewHandler(os.Stderr, reqLoggerOpts))

	e.Pre(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID:    true,
		LogHost:         true,
		LogRemoteIP:     true,
		LogProtocol:     true,
		LogURI:          true,
		LogMethod:       true,
		LogStatus:       true,
		LogLatency:      true,
		LogResponseSize: true,
		LogReferer:      true,
		LogUserAgent:    true,
		LogError:        true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			var userId uint
			user, ok := c.Get("user").(*repo.User)
			if ok && (user != nil) {
				userId = user.Id
			}

			logger.LogAttrs(
				c.Request().Context(),
				slog.LevelInfo,
				"request",
				slog.String("id", v.RequestID),
				slog.String("host", v.Host),
				slog.String("req_ip", v.RemoteIP),
				slog.String("protocol", v.Protocol),
				slog.String("uri", v.URI),
				slog.String("method", v.Method),
				slog.Int("status", v.Status),
				slog.Duration("latency_ms", v.Latency.Round(time.Millisecond)),
				slog.Int64("res_bytes", v.ResponseSize),
				slog.String("referer", v.Referer),
				slog.String("user_agent", v.UserAgent),
				slog.Int("user_id", int(userId)),
				slog.Any("error", v.Error),
			)

			return nil
		},
	}))

	if h.config.RATE_LIMIT_PER_MINUTE > 0 {
		e.Pre(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(h.config.RATE_LIMIT_PER_MINUTE),
				ExpiresIn: time.Minute,
			})}))
	}

	// Gzip compression & decompression

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Skipper: func(c echo.Context) bool {
		return !strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip")
	}}))

	e.Pre(middleware.Decompress())

	pprof.Register(e)

	host, err := os.Hostname()
	if err != nil {
		return nil, errors.Join(errors.New("could not get host name"), err)
	}

	data := map[string]string{
		"env":  h.config.APP_ENV,
		"host": host,
	}

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.tmpl", data)
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	v1 := e.Group("/v1")
	{
		v1.GET("/config", h.GetConfig)
	}

	return e, nil
}
