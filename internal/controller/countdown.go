package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type CountdownData struct {
	Title   string
	Default Default
	Random  Restaurant
}

func (c *Controller) RenderCountdown(ctx echo.Context) error {
	var restaurant Restaurant
	for _, r := range c.Default.FasanenhofRestaurants {
		if r.Selected {
			restaurant = r
			break
		}
	}
	return ctx.Render(http.StatusOK, "countdown", CountdownData{Title: "Einstellungen", Default: c.Default, Random: restaurant})
}
