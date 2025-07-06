package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gitlab.unjx.de/flohoss/mittag/handler"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/services"
)

func setupRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())

	return e
}

func main() {
	e := setupRouter()

	env, err := env.Parse()
	if err != nil {
		e.Logger.Fatal(err.Error())
	}

	e.Logger.SetLevel(env.GetLogLevel())
	if env.GetLogLevel() == log.DEBUG {
		e.Use(middleware.Logger())
		e.Debug = true
	}

	c := services.NewConfigParser()
	r := services.NewMittag(c.Restaurants)
	defer r.Close()

	handler.SetupRouter(e, handler.NewMittagHandler(r), env.APIToken)

	e.Logger.Infof("Server starting on http://localhost:%d", env.Port)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(fmt.Sprintf(":%d", env.Port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	e.Logger.Info("Received shutdown signal. Exiting immediately.")
}
