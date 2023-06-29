package controller

import (
	"os"

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
