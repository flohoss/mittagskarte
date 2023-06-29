package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
)

type CountdownData struct {
	Title      string
	Navigation [][]restaurant.Restaurant
	Random     restaurant.Restaurant
}

func (c *Controller) RenderCountdown(ctx echo.Context) error {
	var restaurant restaurant.Restaurant
	for _, r := range c.Navigation[0] {
		if r.Selected {
			restaurant = r
			break
		}
	}
	return ctx.Render(http.StatusOK, "countdown", CountdownData{Title: "Einstellungen", Navigation: c.Navigation, Random: restaurant})
}
