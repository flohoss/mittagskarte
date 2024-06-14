package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logger"
	"gitlab.unjx.de/flohoss/mittag/internal/router"
	"gitlab.unjx.de/flohoss/mittag/internal/service"
)

func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(logger.NewLogger(env.LogLevel))

	config := config.NewConfig()
	service.NewUpdateService(config)
	handler := router.NewHandler(config.Restaurants)
	router := router.NewRouter(handler)

	slog.Info("starting server", "url", fmt.Sprintf("http://localhost:%d", env.Port))
	if err := router.Echo.Start(fmt.Sprintf(":%d", env.Port)); err != http.ErrServerClosed {
		slog.Error("cannot start server", "err", err)
		os.Exit(1)
	}
}
