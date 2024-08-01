package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/labstack/echo/v4"

	"github.com/rohitxdev/go-api-template/config"
	_ "github.com/rohitxdev/go-api-template/docs"
	"github.com/rohitxdev/go-api-template/embedded"
	"github.com/rohitxdev/go-api-template/handler"
	"github.com/rohitxdev/go-api-template/util"
)

func StartServer(e *echo.Echo) {
	address := config.HOST + ":" + config.PORT

	if config.HTTPS {
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

func main() {
	buildFile, err := embedded.FS.ReadFile("build.json")
	if err != nil {
		panic("could not read build.json file: " + err.Error())
	}
	fmt.Println("BUILD INFO")
	util.PrintTableJSON(buildFile)

	e := handler.GetEcho()

	go StartServer(e)

	util.RegisterCleanUp("server", func() error {
		return e.Shutdown(context.TODO())
	})

	util.SetCleanUp()
}
