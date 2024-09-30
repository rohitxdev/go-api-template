package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rohitxdev/go-api-starter/internal/config"
	"github.com/rohitxdev/go-api-starter/internal/handler"
	"github.com/rohitxdev/go-api-starter/pkg/blobstore"
	"github.com/rohitxdev/go-api-starter/pkg/prettylog"
	"github.com/rohitxdev/go-api-starter/pkg/repo"
	"github.com/rohitxdev/go-api-starter/pkg/sqlite"
	_ "go.uber.org/automaxprocs"
)

// This is set at build time.
var BuildId string

//go:embed web
var staticFS embed.FS

func main() {
	if BuildId == "" {
		panic("build id is not set")
	}

	//Load config
	envFile := flag.String("env-file", ".env", "Path to .env file")
	flag.Parse()

	c, err := config.Load(*envFile)
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

	var logHandler slog.Handler
	if c.Env == config.EnvDevelopment {
		logHandler = prettylog.NewHandler(os.Stderr, &logOpts)
	} else {
		logHandler = slog.NewJSONHandler(os.Stderr, &logOpts)
	}

	slog.SetDefault(slog.New(logHandler))

	slog.Debug(BuildId + " is running in " + c.Env + " environment")

	//Connect to database
	db, err := sql.Open("postgres", c.DatabaseUrl)
	if err != nil {
		panic("connect to database: " + err.Error())
	}
	defer func() {
		if err = db.Close(); err != nil {
			panic("close database: " + err.Error())
		}
		slog.Debug("database connection closed ✔︎")
	}()

	if err = db.Ping(); err != nil {
		panic("ping database: " + err.Error())
	}

	slog.Debug("connected to database ✔︎")

	//Connect to kv store
	kv, err := sqlite.NewKV("kv", sqlite.KVOpts{CleanUpFreq: time.Minute * 5})
	if err != nil {
		panic("connect to kv store: " + err.Error())
	}
	defer func() {
		kv.Close()
		slog.Debug("kv store closed ✔︎")
	}()
	slog.Debug("connected to kv store ✔︎")

	//Create API handler
	r := repo.New(db)
	defer r.Close()

	s3Client, err := blobstore.New(c.S3Endpoint, c.S3DefaultRegion, c.AwsAccessKeyId, c.AwsAccessKeySecret)
	if err != nil {
		panic("connect to s3 client: " + err.Error())
	}

	opts := handler.Opts{
		Config:   c,
		Kv:       kv,
		Repo:     r,
		Email:    nil,
		Fs:       s3Client,
		StaticFS: &staticFS,
	}

	e, err := handler.New(&opts)
	if err != nil {
		panic("create router: " + err.Error())
	}

	//Create tcp listener & start server
	ls, err := net.Listen("tcp", c.Host+":"+c.Port)
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

	go func() {
		if err := http.Serve(ls, e); err != nil && !errors.Is(err, net.ErrClosed) {
			panic("serve http: " + err.Error())
		}
	}()

	slog.Debug("http server started ✔︎")
	slog.Info("server is listening to \x1b[32mhttp://" + ls.Addr().String() + "\x1b[0m and is ready to serve requests ✔︎")

	//Shut down http server gracefully
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), c.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic("http server shutdown: " + err.Error())
	}

	slog.Debug("shut down http server gracefully ✔︎")
}
