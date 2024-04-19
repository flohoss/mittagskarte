package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.unjx.de/flohoss/mittag/internal/controller"
)

func InitRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Debug = true

	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Pre(middleware.RemoveTrailingSlash())

	e.Renderer = initTemplates()

	return e
}

func SetupRoutes(e *echo.Echo, ctrl *controller.Controller, adminKey string) {
	static := e.Group("/static", longCacheLifetime)
	static.Static("/", "web/static")

	public := e.Group("/storage/public", longCacheLifetime)
	public.Static("/", "storage/public")

	modules := e.Group("/modules", longCacheLifetime)
	modules.Static("/", "web/node_modules")

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
