package controller

import (
	"os"

	"github.com/robfig/cron/v3"
	"gitlab.unjx.de/flohoss/mittag/internal/database"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/maps"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
	"gorm.io/gorm"
)

type Controller struct {
	orm        *gorm.DB
	env        *env.Config
	schedule   *cron.Cron
	Navigation [][]restaurant.Restaurant
}

func NewController(env *env.Config) *Controller {
	db := database.NewDatabaseConnection("sqlite.db")
	ctrl := Controller{orm: db, env: env}
	restaurant.MigrateModels(ctrl.orm)
	ctrl.Navigation = restaurant.GetNavigation(ctrl.orm)
	ctrl.createMaps()

	return &ctrl
}

func (c *Controller) setupSchedule() {
	c.schedule = cron.New()

	c.schedule.AddFunc("10 8-14 * * *", func() {
		c.UpdateAllRestaurants()
	})
	c.schedule.AddFunc("0 0 * * *", func() {
		c.setRandomRestaurant()
	})

	c.schedule.Start()
}

func (c *Controller) createMaps() {
	for _, restaurants := range c.Navigation {
		for _, restaurant := range restaurants {
			folder := fetch.DownloadLocation + restaurant.ID
			os.MkdirAll(folder, os.ModePerm)
			if _, err := os.Stat(folder + "/map.webp"); err == nil {
				continue
			}
			maps.CreateMap(restaurant.Latitude, restaurant.Longitude, folder)
		}
	}
}
