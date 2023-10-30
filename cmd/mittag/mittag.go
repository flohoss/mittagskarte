package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"gitlab.unjx.de/flohoss/mittag/internal/controller"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logging"
	"gitlab.unjx.de/flohoss/mittag/internal/router"
)

func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(logging.CreateLogger(env.LogLevel))

	r := router.InitRouter()
	c := controller.NewController(env)
	router.SetupRoutes(r, c, env.AdminKey)

	slog.Info("starting server", "url", fmt.Sprintf("http://localhost:%d", env.Port), "version", env.Version)
	if err := r.Start(fmt.Sprintf(":%d", env.Port)); err != http.ErrServerClosed {
		slog.Error("cannot start server", "err", err)
		os.Exit(1)
	}
}
