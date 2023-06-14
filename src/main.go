package main

import (
	"fmt"
	"mittag/controller"
	"mittag/env"
	"net/http"
)

func main() {
	env := env.Parse()

	logger := setupLogger(env.LogLevel)
	defer logger.Sync()

	controller := controller.NewController(env, logger)
	router := initRouter()
	setupRoutes(router, controller, env.AdminKey)

	logger.Infow("Starting server", "url", fmt.Sprintf("http://localhost:%d", env.Port))
	if err := router.Start(fmt.Sprintf(":%d", env.Port)); err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}
