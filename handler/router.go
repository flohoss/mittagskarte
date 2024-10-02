package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "gitlab.unjx.de/flohoss/mittag/docs"
)

type Router struct {
	Echo       *echo.Echo
	handler    *MittagHandler
	formAuth   echo.MiddlewareFunc
	bearerAuth echo.MiddlewareFunc
}

func NewRouter(handler *MittagHandler, token string) *Router {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().URL.Path, "docs")
		},
	}))

	r := &Router{
		Echo:    e,
		handler: handler,
		bearerAuth: middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
			return key == token, nil
		}),
	}
	r.SetupRoutes()
	return r
}

func longCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000")
		return next(c)
	}
}

func (r *Router) SetupRoutes() {
	r.Echo.GET("/api/docs/*", echoSwagger.WrapHandler)
	r.Echo.GET("api/docs", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/api/docs/index.html")
	})

	api := r.Echo.Group("/api/v1")
	api.GET("/restaurants", r.handler.GetAllRestaurants)
	api.GET("/restaurants/:id", r.handler.GetRestaurant)
	api.POST("/restaurants/:id", r.handler.UploadMenu, r.bearerAuth)

	r.Echo.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	public := r.Echo.Group("/data", longCacheLifetime)
	public.Static("/thumbnails", "data/thumbnails")

	r.Echo.Static("/storage/downloads", "storage/downloads")
	r.Echo.Static("/assets", "web/assets")
	r.Echo.Static("/favicon", "web/favicon")
	r.Echo.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "index.html", nil)
	})
}
