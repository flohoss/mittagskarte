package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"gitlab.unjx.de/flohoss/mittag/docs"
	"gitlab.unjx.de/flohoss/mittag/handler"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logger"
	"gitlab.unjx.de/flohoss/mittag/services"
)

// @title			Mittagstisch API
// @version		1.0
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath		/api/v1
func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(logger.New(env.LogLevel))

	publicUrl, _ := url.Parse(env.PublicUrl)
	docs.SwaggerInfo.Host = publicUrl.Host
	docs.SwaggerInfo.Schemes = []string{publicUrl.Scheme}

	c := services.NewConfigParser()
	r := services.NewMittag(c.Restaurants)
	defer r.Close()

	router := handler.NewRouter(handler.NewMittagHandler(r), env.APIToken)
	slog.Info("server listening, press ctrl+c to stop", "addr", env.PublicUrl)
	if err := router.Echo.Start(fmt.Sprintf(":%d", env.Port)); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server terminated", "error", err)
		os.Exit(1)
	}
}
