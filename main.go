package main

import (
	"fmt"
	"log/slog"
	"os"

	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/crawl"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logger"
	"gitlab.unjx.de/flohoss/mittag/internal/parse"
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
		crawl := crawl.NewCrawler(config.PageURL, config.Parse.HTTPVersion, config.Parse.Navigate, config.Parse.IsFile)
		if config.Parse.IsFile {
			parser := parse.NewParser(config.ID, crawl.FinalUrl, config.Parse.HTTPVersion)
			if parser.Content != "" {
				fmt.Println(parser.DownloadedFile)
			}
		}
	}
}
