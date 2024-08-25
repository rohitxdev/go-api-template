package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rohitxdev/go-api-template/pkg/config"
	"github.com/rohitxdev/go-api-template/pkg/handler"
	"github.com/rohitxdev/go-api-template/pkg/repo"
)

//go:embed templates public
var staticFS embed.FS

func main() {
	c, err := config.LoadConfig(".env")
	if err != nil {
		panic("could not load config: " + err.Error())
	}

	db, err := sql.Open("postgres", c.DB_URL)
	if err != nil {
		panic("could not connect to PostgreSQL database: " + err.Error())
	}
	defer db.Close()

	r := repo.NewRepo(db)
	h := handler.NewHandler(c, r, &staticFS)

	e, err := handler.NewRouter(h)
	if err != nil {
		panic("could not create router: " + err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := e.Start(fmt.Sprintf("%s:%s", c.HOST, c.PORT)); !errors.Is(err, http.ErrServerClosed) && err != nil {
			panic("could not start HTTP server: " + err.Error())
		}
	}()

	<-ctx.Done()

	fmt.Println("\nShutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		panic("could not shutdown server gracefully: " + err.Error())
	}

	wg.Wait()
	fmt.Println("Server shutdown gracefully")
}
