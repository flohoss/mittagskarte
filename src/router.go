package main

import (
	"mittag/controller"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func initRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	e.Renderer = initTemplates()

	return e
}

func longCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000")
		return next(c)
	}
}

func setupRoutes(e *echo.Echo, ctrl *controller.Controller, adminKey string) {
	static := e.Group("/static", longCacheLifetime)
	static.Static("/", "static")

	e.GET("/countdown", ctrl.RenderCountdown)
	e.GET("/settings", ctrl.RenderSettings)

	restaurants := e.Group("/restaurants")
	restaurants.GET("/:id", ctrl.RenderRestaurants)
	restaurants.PATCH("", ctrl.UpdateRestaurants, middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == adminKey, nil
	}))

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})
	e.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/countdown")
	})
}
