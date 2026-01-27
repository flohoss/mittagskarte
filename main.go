package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/handlers"
	"github.com/flohoss/mittagskarte/services"
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

func initSentry(sentryDSN string) {
	if sentryDSN == "" {
		slog.Warn("Sentry DSN is empty, skipping Sentry initialization")
		return
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
	})
	if err != nil {
		slog.Error("sentry.Init", "error", err)
		os.Exit(1)
	}
}

func main() {
	e := setupRouter()
	config.New()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.GetLogLevel(),
	}))
	slog.SetDefault(logger)

	m := services.NewMittag(config.GetRestaurants())
	defer m.Close()

	mh := handlers.NewMittagHandler(m)

	handlers.SetupRouter(e, mh)

	slog.Info("Starting server", "url", fmt.Sprintf("http://%s", config.GetServer()))
	slog.Error(e.Start(config.GetServer()).Error())
}
