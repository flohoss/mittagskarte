package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/mittag"
)

type Controller struct {
	env    *env.Env
	mittag *mittag.Mittag
	cron   *cron.Cron
}

func NewController(env *env.Env) *Controller {
	ctrl := new(Controller)

	ctrl.env = env
	ctrl.mittag = mittag.NewMittag()
	ctrl.cron = cron.New()
	ctrl.cron.AddFunc("0,30 10,11 * * *", ctrl.updateAll)
	ctrl.cron.Start()

	return ctrl
}

func (c *Controller) updateAll() {
	c.mittag.UpdateRestaurants()
}

func (c *Controller) UpdateRestaurants(ctx echo.Context) error {
	id := ctx.QueryParam("id")
	if id != "" {
		exists, conf := c.mittag.DoesConfigurationExist(id)
		if !exists {
			return ctx.NoContent(http.StatusNotFound)
		}
		go func() {
			conf.UpdateInformation(c.mittag.GetORM())
		}()
	} else {
		go c.mittag.UpdateRestaurants()
	}
	return ctx.NoContent(http.StatusOK)
}
