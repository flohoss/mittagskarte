package controller

import (
	"mittag/database"
	"mittag/env"
	"mittag/fetch"
	"mittag/maps"
	"os"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Controller struct {
	orm      *gorm.DB
	env      *env.Config
	log      *zap.SugaredLogger
	schedule *cron.Cron
	Default  Default
}

type Default struct {
	StuttgartRestaurants  []Restaurant
	NuertingenRestaurants []Restaurant
}

func NewController(env *env.Config, logger *zap.SugaredLogger) *Controller {
	db := database.NewDatabaseConnection("sqlite.db")

	ctrl := Controller{orm: db, env: env, log: logger}
	ctrl.setupSchedule()
	ctrl.MigrateModels()
	ctrl.setupDefaults()
	ctrl.createMaps()

	return &ctrl
}

func (c *Controller) setupDefaults() {
	c.orm.Where("City IN ?", []string{"Leinfelden-Echterdingen", "Stuttgart"}).Select("ID", "Name", "Selected").Order("Name").Find(&c.Default.StuttgartRestaurants)
	c.orm.Where("City IN ?", []string{"Nürtingen", "Oberboihingen"}).Select("ID", "Name", "Selected").Order("Name").Find(&c.Default.NuertingenRestaurants)
}

func (c *Controller) setupSchedule() {
	c.schedule = cron.New()

	c.schedule.AddFunc("10 8-14 * * *", func() {
		c.updateData()
	})
	c.schedule.AddFunc("0 0 * * *", func() {
		c.setRandomRestaurant()
	})

	c.schedule.Start()
}

func (c *Controller) updateData() {
	restaurants := c.restaurantsJoinCardJoinFood()
	for _, restaurant := range restaurants {
		c.updateRestaurantData(&restaurant)
	}
}

func (c *Controller) createMaps() {
	restaurants := c.restaurants()
	for _, restaurant := range restaurants {
		folder := fetch.DownloadLocation + restaurant.ID
		os.MkdirAll(folder, os.ModePerm)
		if _, err := os.Stat(folder + "/map.webp"); err == nil {
			continue
		}
		maps.CreateMap(restaurant.Latitude, restaurant.Longitude, folder)
	}
}
