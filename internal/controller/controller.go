package controller

import (
	"os"

	"github.com/robfig/cron/v3"
	"gitlab.unjx.de/flohoss/mittag/internal/database"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/fetch"
	"gitlab.unjx.de/flohoss/mittag/internal/maps"
	"gorm.io/gorm"
)

type Controller struct {
	orm      *gorm.DB
	env      *env.Config
	schedule *cron.Cron
	Default  Default
}

type Default struct {
	FasanenhofRestaurants []Restaurant
	FlorianRestaurants    []Restaurant
	AndreRestaurants      []Restaurant
}

func NewController(env *env.Config) *Controller {
	db := database.NewDatabaseConnection("sqlite.db")

	ctrl := Controller{orm: db, env: env}
	ctrl.setupSchedule()
	ctrl.MigrateModels()
	ctrl.setupDefaults()
	ctrl.createMaps()

	return &ctrl
}

func (c *Controller) setupDefaults() {
	c.orm.Where(&Restaurant{Group: Fasanenhof}).Select("ID", "Name", "Selected").Order("Name").Find(&c.Default.FasanenhofRestaurants)
	c.orm.Where(&Restaurant{Group: Florian}).Select("ID", "Name", "Selected").Order("Name").Find(&c.Default.FlorianRestaurants)
	c.orm.Where(&Restaurant{Group: Andre}).Select("ID", "Name", "Selected").Order("Name").Find(&c.Default.AndreRestaurants)
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
