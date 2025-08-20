package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.unjx.de/flohoss/mittag/config"
	"gitlab.unjx.de/flohoss/mittag/handlers"
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
