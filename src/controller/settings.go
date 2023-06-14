package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SystemData struct {
	Title   string
	Default Default
}

func (c *Controller) RenderSettings(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "settings", SystemData{Title: "Einstellungen", Default: c.Default})
}
