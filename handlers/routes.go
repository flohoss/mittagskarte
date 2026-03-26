package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/flohoss/mittagskarte/internal/events"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r3labs/sse/v2"
)

func longCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000")
		return next(c)
	}
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, ".")
}

func InitRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "events")
		},
	}))

	e.Renderer = initTemplates()

	return e
}

func SetupRouter(e *echo.Echo, mh *RestaurantHandler) {
	e.GET("/health", healthHandler)
	e.HEAD("/health", healthHandler)

	h := huma.DefaultConfig("Mittagskarte API", os.Getenv("APP_VERSION"))
	h.OpenAPIPath = "/api/openapi"
	h.DocsPath = "/api/docs"
	h.SchemasPath = "/api/schemas"
	humaAPI := humaecho.New(e, h)

	huma.Register(humaAPI, mh.listRestaurantsOperation(), mh.listRestaurantsHandler)
	huma.Register(humaAPI, mh.getRestaurantOperation(), mh.getRestaurantHandler)

	mh.mittag.SetEvents(events.New(func(streamID string, sub *sse.Subscriber) {
		mh.mittag.Events.SendUpdate(mh.listRestaurants())
	}))
	e.GET("/api/events", mh.mittag.Events.GetHandler())

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	assets := e.Group("/assets", longCacheLifetime)
	assets.Static("/", "web/assets")

	favicon := e.Group("/static", longCacheLifetime)
	favicon.Static("/", "web/static")

	thumbnails := e.Group("/thumbnails", longCacheLifetime)
	thumbnails.Static("/", "config/thumbnails")

	downloads := e.Group("/config/downloads", longCacheLifetime)
	downloads.Static("/", "config/downloads")

	e.Any("/*", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "index.html", nil)
	})
}
