package controller

import (
	"github.com/robfig/cron/v3"
	"gitlab.unjx.de/flohoss/mittag/internal/database"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/maps"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
	"gorm.io/gorm"
)

type Controller struct {
	orm            *gorm.DB
	env            *env.Config
	schedule       *cron.Cron
	Navigation     [][]restaurant.Restaurant
	MapInformation map[string]*maps.MapInformation
}

func NewController(env *env.Config) *Controller {
	db := database.NewDatabaseConnection("sqlite.db")
	ctrl := Controller{orm: db, env: env}
	restaurant.MigrateModels(ctrl.orm)
	ctrl.Navigation = restaurant.GetNavigation(ctrl.orm)
	ctrl.createMaps()
	ctrl.setupSchedule()

	return &ctrl
}

func (c *Controller) setupSchedule() {
	c.schedule = cron.New()

	c.schedule.AddFunc("0,30 10,11 * * *", func() {
		c.UpdateAllRestaurants()
	})
	c.schedule.AddFunc("0 0 * * *", func() {
		c.setRandomRestaurant()
	})

	c.schedule.Start()
}

func (c *Controller) createMaps() {
	mapRequests := []maps.MapRequest{}
	for _, restaurants := range c.Navigation {
		for _, restaurant := range restaurants {
			mapRequests = append(mapRequests, maps.MapRequest{
				Identifier: restaurant.ID,
				Address:    restaurant.Address,
			})
		}
	}
	c.MapInformation = maps.GetMapInformation(c.env.GoogleAPIKey, mapRequests)
}
