package main

import (
	"log/slog"
	"os"

	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logger"
	"gitlab.unjx.de/flohoss/mittag/services"
)

// @title			Mittagstisch API
// @version		1.0
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @host			mittag.unjx.de
// @schemes		https
// @BasePath		/api/v1
func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(logger.New(env.LogLevel))

	c := services.NewConfigParser()
	r := services.NewRestaurantHandler(c.Restaurants)
	defer r.Close()
}
