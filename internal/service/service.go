package service

import (
	"github.com/robfig/cron/v3"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/crawl"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/imdb"
	"gitlab.unjx.de/flohoss/mittag/internal/parse"
)

type UpdateService struct {
	cron   *cron.Cron
	config *config.Config
	imdb   *imdb.IMDb
	env    *env.Env
}

func New(config *config.Config, imdb *imdb.IMDb, env *env.Env) *UpdateService {
	u := &UpdateService{
		cron:   cron.New(),
		config: config,
		imdb:   imdb,
		env:    env,
	}
	u.RestoreMenus()
	u.cron.AddFunc("0,30 10,11 * * *", u.updateAll)
	u.updateAll()
	u.cron.Start()
	return u
}

func (u *UpdateService) RestoreMenus() {
	for _, config := range u.config.Restaurants {
		config.RestoreMenu(u.imdb)
	}
}

func (u *UpdateService) updateAll() {
	for _, config := range u.config.Restaurants {
		u.UpdateSingle(config)
	}
}

func (u *UpdateService) UpdateSingle(restaurant *config.Restaurant) {
	crawl := crawl.NewCrawler(restaurant.PageURL, restaurant.Parse.HTTPVersion, restaurant.Parse.Navigate, restaurant.Parse.IsFile)
	var fileContent, card string
	if restaurant.Parse.IsFile {
		needsParsing := restaurant.Parse.OneForAll.Regex != "" || len(restaurant.Parse.Food) > 0
		p := parse.NewFileParser(restaurant.ID, crawl.FinalUrl, restaurant.Parse.HTTPVersion, needsParsing)
		card = p.OutputFileLocation
		fileContent = p.OutputFileContent
	}
	parser := parse.NewMenuParser(crawl.DocStorage, fileContent, &restaurant.Parse, card)
	restaurant.Menu = *parser.Menu
	restaurant.SaveMenu(u.imdb)
}
