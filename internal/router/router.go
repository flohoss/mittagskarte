package router

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "gitlab.unjx.de/flohoss/mittag/docs"
	"gitlab.unjx.de/flohoss/mittag/internal/handler"
)

type Router struct {
	Echo    *echo.Echo
	handler *handler.RestaurantHandler
}

func New(handler *handler.RestaurantHandler) *Router {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().URL.Path, "docs")
		},
	}))
	e.Pre(middleware.RemoveTrailingSlash())

	r := &Router{
		Echo:    e,
		handler: handler,
	}
	r.SetupRoutes()
	return r
}

func (r *Router) SetupRoutes() {
	public := r.Echo.Group("/public", longCacheLifetime)
	public.Static("/menus", "storage/public/menus")
	public.Static("/thumbnails", "internal/config/thumbnails")

	r.Echo.GET("/api/docs/*", echoSwagger.WrapHandler)
	r.Echo.GET("api/docs", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/api/docs/index.html")
	})

	api := r.Echo.Group("/api/v1")
	api.GET("/groups", r.handler.GetAllRestaurantsGrouped)
	api.GET("/restaurants", r.handler.GetAllRestaurants)
	api.GET("/restaurants/:id", r.handler.GetRestaurant)

	r.Echo.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})
	r.Echo.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/api/docs/index.html")
	})
}
