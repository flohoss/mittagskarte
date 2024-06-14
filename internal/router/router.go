package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Router struct {
	Echo    *echo.Echo
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Pre(middleware.RemoveTrailingSlash())

	r := &Router{
		Echo:    e,
		handler: handler,
	}
	r.SetupRoutes()
	return r
}

func (r *Router) SetupRoutes() {
	public := r.Echo.Group("/storage/public", longCacheLifetime)
	public.Static("/", "storage/public")

	api := r.Echo.Group("/api/v1")
	api.GET("/restaurants", r.handler.GetAllRestaurants)
	api.GET("/restaurants/:id", r.handler.GetRestaurant)

	r.Echo.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})
	r.Echo.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/docs/swagger")
	})
}
