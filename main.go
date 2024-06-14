package main

import (
	"fmt"
	"log/slog"
	"os"

	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/crawl"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logger"
)

func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(logger.NewLogger(env.LogLevel))

	config := config.NewConfig()
	for _, config := range config.Restaurants {
		crawl.NewCrawler(config.PageURL, config.Parse.HTTPVersion, config.Parse.Navigate, config.Parse.IsFile)
	}
	fmt.Println("done")
}
