package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "gitlab.unjx.de/flohoss/mittag/docs"
	"golang.org/x/time/rate"
)

type Router struct {
	Echo        *echo.Echo
	handler     *MittagHandler
	bearerAuth  echo.MiddlewareFunc
	rateLimiter echo.MiddlewareFunc
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
	e.Renderer = initTemplates()

	timeout := 12 * time.Hour
	r := &Router{
		Echo:    e,
		handler: handler,
		bearerAuth: middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, c echo.Context) (bool, error) {
				return key == token, nil
			},
			ErrorHandler: func(err error, c echo.Context) error {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentifizierung fehlgeschlagen")
			},
		}),
		rateLimiter: middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(1 / timeout.Seconds()), Burst: 1, ExpiresIn: timeout},
			),
			IdentifierExtractor: func(ctx echo.Context) (string, error) {
				restaurant := ctx.Param("id")
				if restaurant == "" {
					restaurant = "none"
				}
				return restaurant, nil
			},
			ErrorHandler: func(context echo.Context, err error) error {
				return echo.NewHTTPError(http.StatusForbidden, nil)
			},
			DenyHandler: func(context echo.Context, identifier string, err error) error {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Zu viele Anfragen für dieses Restaurant (max 1/12h). Bitte versuchen Sie es in ein paar Minuten erneut.")
			},
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
	api.PUT("/restaurants/:id", r.handler.RefreshRestaurant, r.rateLimiter)
	api.POST("/restaurants/:id", r.handler.UploadMenu, r.bearerAuth)

	r.Echo.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	public := r.Echo.Group("/data", longCacheLifetime)
	public.Static("/thumbnails", "data/thumbnails")

	assets := r.Echo.Group("/assets", longCacheLifetime)
	assets.Static("/", "web/assets")

	favicon := r.Echo.Group("/favicon", longCacheLifetime)
	favicon.Static("/", "web/favicon")

	storage := r.Echo.Group("/storage", longCacheLifetime)
	storage.Static("/downloads", "storage/downloads")

	r.Echo.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "index.html", nil)
	})
}
