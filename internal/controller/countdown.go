package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
)

type CountdownData struct {
	Title      string
	Navigation [][]restaurant.Restaurant
	Random     []restaurant.Restaurant
}

func (c *Controller) RenderCountdown(ctx echo.Context) error {
	var random []restaurant.Restaurant
	for i, _ := range restaurant.Groups {
		for _, r := range c.Navigation[i] {
			if r.Selected {
				random = append(random, r)
				break
			}
		}
	}
	return ctx.Render(http.StatusOK, "countdown", CountdownData{Title: "Mittag - Countdown", Navigation: c.Navigation, Random: random})
}
