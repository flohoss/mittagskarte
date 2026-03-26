package main

import (
	"log/slog"
	"os"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/handlers"
	"github.com/flohoss/mittagskarte/services"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.GetLogLevel(),
	}))
	slog.SetDefault(logger)

	m := services.NewMittag()
	defer m.Close()

	mh := handlers.NewRestaurantHandler(m)

	e := handlers.InitRouter()
	handlers.SetupRouter(e, mh)

	slog.Info("Starting server", "url", config.GetServerURL())
	slog.Error(e.Start(config.GetServer()).Error())
}
