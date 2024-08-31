package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/handler"
	"github.com/rohitxdev/go-api-template/pkg/logger"
	"github.com/rohitxdev/go-api-template/pkg/repo"
)

// Build info
var BuildInfo string

//go:embed templates public
var staticFS embed.FS

func main() {
	defer fmt.Println()

	//Load config
	cfg, err := config.Load(".env")
	if err != nil {
		panic("load config: " + err.Error())
	}

	//Set up logger
	loggerOpts := logger.HandlerOpts{
		TimeFormat: time.RFC3339,
		Level:      slog.LevelDebug,
		NoColor:    !cfg.IS_DEV,
	}

	if cfg.IS_DEV {
		loggerOpts.TimeFormat = time.Kitchen
	}

	slog.SetDefault(slog.New(logger.NewHandler(os.Stderr, loggerOpts)))

	slog.Debug("running " + BuildInfo + " in " + cfg.APP_ENV + " environment")

	//Connect to database
	slog.Debug("connecting to database... ⏳")

	db, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		panic("connect to database: " + err.Error())
	}
	defer func() {
		if err = db.Close(); err != nil {
			panic("close database: " + err.Error())
		}
		slog.Debug("database connection closed ✔︎")
	}()

	slog.Debug("connected to database ✔︎")

	//Create handler
	r := repo.New(db)
	h := handler.New(cfg, r, &staticFS)

	e, err := handler.NewRouter(h)
	if err != nil {
		panic("create router: " + err.Error())
	}

	//Create tcp listener
	slog.Debug("creating tcp listener... ⏳")

	ls, err := net.Listen("tcp", cfg.HOST+":"+cfg.PORT)
	if err != nil {
		panic("tcp listen: " + err.Error())
	}
	defer func() {
		if err = ls.Close(); err != nil {
			panic("close tcp listener: " + err.Error())
		}
		slog.Debug("tcp listener closed ✔︎")
	}()
	slog.Debug("tcp listener created ✔︎")

	slog.Debug("http server started ✔︎")
	slog.Info("server is listening to \033[32mhttp://" + ls.Addr().String() + "\033[0m")

	go func() {
		if err := http.Serve(ls, e); err != nil && !errors.Is(err, net.ErrClosed) {
			panic("serve http: " + err.Error())
		}
	}()

	//Graceful shutdown

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	<-ctx.Done()

	slog.Debug("shutting down http server... ⏳")

	ctx, cancel = context.WithTimeout(context.Background(), cfg.SHUTDOWN_TIMEOUT)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic("http server shutdown: " + err.Error())
	}

	slog.Debug("server shut down gracefully ✔︎")
}
