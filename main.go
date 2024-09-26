package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"gitlab.unjx.de/flohoss/mittag/handler"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logger"
	"gitlab.unjx.de/flohoss/mittag/services"
)

//	@title			Mittagstisch API
//	@version		1.0
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@host			mittag.unjx.de
//	@schemes		https
//	@BasePath		/api/v1
func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(logger.New(env.LogLevel))

	c := services.NewConfigParser()
	r := services.NewMittag(c.Restaurants)
	defer r.Close()

	router := handler.NewRouter(handler.NewMittagHandler(r), env.APIToken)
	slog.Info("starting server", "url", fmt.Sprintf("http://localhost:%d", env.Port))
	if err := router.Echo.Start(fmt.Sprintf(":%d", env.Port)); err != http.ErrServerClosed {
		slog.Error("cannot start server", "err", err)
		os.Exit(1)
	}
}
