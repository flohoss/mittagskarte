package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/handlers"
	"github.com/flohoss/mittagskarte/services"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func setupRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())

	e.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))

	return e
}

func initSentry() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://773c98ff883e064b35dcd522404bac0d@o4510702473641984.ingest.de.sentry.io/4510702475411536",
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
}

func main() {
	initSentry()

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
