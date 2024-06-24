package router

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "gitlab.unjx.de/flohoss/mittag/docs"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/handler"
)

type Router struct {
	Echo       *echo.Echo
	handler    *handler.RestaurantHandler
	formAuth   echo.MiddlewareFunc
	bearerAuth echo.MiddlewareFunc
}

func New(handler *handler.RestaurantHandler, env *env.Env) *Router {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: env.AllowedHosts,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().URL.Path, "docs")
		},
	}))
	e.Renderer = initTemplates()

	r := &Router{
		Echo:    e,
		handler: handler,
		formAuth: middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			KeyLookup: "form:token",
			Validator: func(key string, c echo.Context) (bool, error) {
				return key == env.APIToken, nil
			},
		}),
		bearerAuth: middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
			return key == env.APIToken, nil
		}),
	}
	r.SetupRoutes()
	return r
}

func (r *Router) SetupRoutes() {
	r.Echo.GET("/api/docs/*", echoSwagger.WrapHandler)
	r.Echo.GET("api/docs", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/api/docs/index.html")
	})

	api := r.Echo.Group("/api/v1")
	api.GET("/restaurants", r.handler.GetAllRestaurants)
	api.PATCH("/restaurants", r.handler.UpdateRestaurant, r.bearerAuth)
	api.GET("/restaurants/:id", r.handler.GetRestaurant)
	api.POST("/restaurants/:id", r.handler.UploadMenu, r.formAuth)

	r.Echo.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	public := r.Echo.Group("/config", longCacheLifetime)
	public.Static("/thumbnails", "internal/config/thumbnails")

	r.Echo.Static("/storage/menus", "storage/menus")
	r.Echo.Static("/assets", "web/assets")
	r.Echo.Static("/favicon", "web/favicon")
	r.Echo.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "index.html", nil)
	})
}
