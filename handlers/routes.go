package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func longCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000")
		return next(c)
	}
}

func render(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func SetupRouter(e *echo.Echo) {
	assets := e.Group("/assets", longCacheLifetime)
	assets.Static("/", "assets")

	thumbnails := e.Group("/thumbnails", longCacheLifetime)
	thumbnails.Static("/", "config/thumbnails")

	downloads := e.Group("/config/downloads", longCacheLifetime)
	downloads.Static("/", "config/downloads")

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	e.RouteNotFound("*", handleIndex)
}
