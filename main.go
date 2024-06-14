package main

import (
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
		var fileContent, card string
		if config.Parse.IsFile {
			parse := parse.NewFileParser(config.ID, crawl.FinalUrl, config.Parse.HTTPVersion)
			if !parse.IsNew {
				continue
			}
			card = parse.DownloadedFile
			fileContent = parse.FileContent
		}
		parser := parse.NewMenuParser(crawl.DocStorage, fileContent, &config.Parse, card)
		if len(parser.Menu.Food) > 0 {
			config.Menu = *parser.Menu
			slog.Info("found new menu", "restaurant", config.Name, "card", config.Menu.Card)
		}
	}
}
