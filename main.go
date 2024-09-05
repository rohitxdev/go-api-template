package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/rohitxdev/go-api-template/internal/api"
	"github.com/rohitxdev/go-api-template/internal/config"
	"github.com/rohitxdev/go-api-template/pkg/prettylog"
	"github.com/rohitxdev/go-api-template/pkg/repo"
	"github.com/rohitxdev/go-api-template/pkg/sqlite"
)

//go:embed web
var staticFS embed.FS

func main() {
	//Load config
	cfg, err := config.Load(".env")
	if err != nil {
		panic("load config: " + err.Error())
	}

	//Set up logger
	logOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Value.String() == "" || a.Value.Equal(slog.AnyValue(nil)) {
				return slog.Attr{}
			}
			return a
		},
	}

	var logHandler slog.Handler = slog.NewJSONHandler(os.Stderr, &logOpts)
	if cfg.IsDev {
		logHandler = prettylog.NewHandler(os.Stderr, &logOpts)
	}

	slog.SetDefault(slog.New(logHandler))

	slog.Debug(cfg.BuildInfo + " is running in " + cfg.Env + " environment")

	//Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseUrl)
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

	//Create kv store
	kv, err := sqlite.NewKV("kv")
	if err != nil {
		panic("create kv store: " + err.Error())
	}
	r := repo.New(db)

	//Create API handler
	opts := api.HandlerOpts{
		Config:   cfg,
		Kv:       kv,
		Repo:     r,
		Email:    nil,
		Fs:       nil,
		StaticFS: &staticFS,
	}
	h := api.New(&opts)

	e, err := api.NewRouter(h)
	if err != nil {
		panic("create router: " + err.Error())
	}

	//Create tcp listener
	ls, err := net.Listen("tcp", cfg.Host+":"+cfg.Port)
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
	slog.Info("server is listening to \x1b[32mhttp://" + ls.Addr().String() + "\x1b[0m")

	go func() {
		if err := http.Serve(ls, e); err != nil && !errors.Is(err, net.ErrClosed) {
			panic("serve http: " + err.Error())
		}
	}()

	//Shut down http server gracefully
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic("http server shutdown: " + err.Error())
	}

	slog.Debug("shut down http server gracefully ✔︎")
}
