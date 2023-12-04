package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
	"github.com/vorlif/spreak/humanize"
	"github.com/vorlif/spreak/humanize/locale/de"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/mittag"
	"golang.org/x/text/language"
)

type Controller struct {
	env       *env.Env
	mittag    *mittag.Mittag
	cron      *cron.Cron
	humanizer *humanize.Humanizer
}

func NewController(env *env.Env) *Controller {
	ctrl := new(Controller)

	ctrl.env = env
	collection := humanize.MustNew(humanize.WithLocale(de.New()))
	ctrl.humanizer = collection.CreateHumanizer(language.German)
	ctrl.mittag = mittag.NewMittag(env, ctrl.humanizer)
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

func (c *Controller) UpdateMaps(ctx echo.Context) error {
	id := ctx.QueryParam("id")
	if id != "" {
		exists, _ := c.mittag.DoesConfigurationExist(id)
		if !exists {
			return ctx.NoContent(http.StatusNotFound)
		}
	}
	go c.mittag.UpdateMapsInformation(id)
	return ctx.NoContent(http.StatusOK)
}
