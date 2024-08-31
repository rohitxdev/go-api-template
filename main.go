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

	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/handler"
	"github.com/rohitxdev/go-api-template/pkg/prettylog"
	"github.com/rohitxdev/go-api-template/pkg/repo"
)

// Build info
var BuildInfo string

//go:embed templates public
var staticFS embed.FS

func main() {
	fmt.Println()

	//Load config
	cfg, err := config.Load(".env")
	if err != nil {
		panic("load config: " + err.Error())
	}

	logOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	var logHandler slog.Handler = slog.NewJSONHandler(os.Stderr, &logOpts)
	if cfg.IS_DEV {
		logHandler = prettylog.NewHandler(os.Stderr, &logOpts)
	}

	slog.SetDefault(slog.New(logHandler))

	slog.Debug("running " + BuildInfo + " in " + cfg.APP_ENV + " environment")

	//Connect to database
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
	slog.Info("server is listening to http://" + ls.Addr().String())

	go func() {
		if err := http.Serve(ls, e); err != nil && !errors.Is(err, net.ErrClosed) {
			panic("serve http: " + err.Error())
		}
	}()

	//Shut down http server gracefully

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), cfg.SHUTDOWN_TIMEOUT)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic("http server shutdown: " + err.Error())
	}

	slog.Debug("shut down http server gracefully ✔︎")
}
