package service

import (
	"log/slog"

	"github.com/robfig/cron/v3"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/crawl"
	"gitlab.unjx.de/flohoss/mittag/internal/imdb"
	"gitlab.unjx.de/flohoss/mittag/internal/parse"
)

type UpdateService struct {
	cron   *cron.Cron
	config *config.Config
	imdb   *imdb.IMDb
}

func New(config *config.Config, imdb *imdb.IMDb) *UpdateService {
	u := &UpdateService{
		cron:   cron.New(),
		config: config,
		imdb:   imdb,
	}
	u.RestoreMenus()
	u.cron.AddFunc("0,30 10,11 * * *", u.updateAll)
	//u.updateAll()
	u.cron.Start()
	return u
}

func (u *UpdateService) RestoreMenus() {
	for _, config := range u.config.Restaurants {
		u.imdb.RestoreMenu(config)
	}
}

func (u *UpdateService) updateAll() {
	for _, config := range u.config.Restaurants {
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
			u.imdb.SaveMenu(config)
			slog.Info("found new menu", "restaurant", config.Name, "card", config.Menu.Card)
		}
	}
}
