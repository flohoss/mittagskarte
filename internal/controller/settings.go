package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
)

type SystemData struct {
	Title      string
	Navigation [][]restaurant.Restaurant
}

func (c *Controller) RenderSettings(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "settings", SystemData{Title: "Mittag - Einstellungen", Navigation: c.Navigation})
}
