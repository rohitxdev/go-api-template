package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/rohitxdev/go-api-template/config"
	"github.com/rohitxdev/go-api-template/embedded"
	"github.com/rohitxdev/go-api-template/handler"
	"github.com/rohitxdev/go-api-template/repo"
	"github.com/rohitxdev/go-api-template/util"
)

func main() {
	c, err := config.LoadConfig(".env")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", c.DB_URL)
	if err != nil {
		panic("could not connect to PostgreSQL database: " + err.Error())
	}
	defer db.Close()

	r := repo.NewRepo(db)
	h := handler.NewHandler(c, r)
	e, err := handler.InitRouter(h)
	if err != nil {
		panic(err)
	}

	buildFile, err := embedded.FS.ReadFile("build.json")
	if err != nil {
		panic("could not read build.json file: " + err.Error())
	}
	fmt.Println("BUILD INFO")
	util.PrintTableJSON(buildFile)

	address := c.HOST + ":" + c.PORT

	if c.HTTPS {
		certFile, err := embedded.FS.ReadFile("certs/cert.pem")
		if err != nil {
			panic("could not read cert file:" + err.Error())
		}
		keyFile, err := embedded.FS.ReadFile("certs/key.pem")
		if err != nil {
			panic("could not read key file: " + err.Error())
		}
		if err := e.StartTLS(address, certFile, keyFile); !errors.Is(err, http.ErrServerClosed) && err != nil {
			panic("could not start HTTPS server: " + err.Error())
		}
	} else {
		if err := e.Start(address); !errors.Is(err, http.ErrServerClosed) && err != nil {
			panic("could not start HTTP server: " + err.Error())
		}
	}

}
