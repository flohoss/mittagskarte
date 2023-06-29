package router

import (
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.unjx.de/flohoss/mittag/internal/controller"
)

func InitRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(echo.WrapMiddleware(chiMiddleware.Heartbeat("/health")))

	e.Renderer = initTemplates()

	return e
}

func SetupRoutes(e *echo.Echo, ctrl *controller.Controller, adminKey string) {
	static := e.Group("/static", longCacheLifetime)
	static.Static("/", "web/static")

	storage := e.Group("/storage/downloads", longCacheLifetime)
	storage.Static("/", "storage/downloads")

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
