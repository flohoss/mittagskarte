package handlers

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.unjx.de/flohoss/mittag/config"
	"golang.org/x/time/rate"
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

func SetupRouter(e *echo.Echo, mh *MittagHandler) {
	assets := e.Group("/assets", longCacheLifetime)
	assets.Static("/", "assets")

	thumbnails := e.Group("/thumbnails", longCacheLifetime)
	thumbnails.Static("/", "config/thumbnails")

	downloads := e.Group("/config/downloads", longCacheLifetime)
	downloads.Static("/", "config/downloads")

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	e.GET("/filter", mh.handleFilter)
	e.POST("/upload/:id", mh.handleUpload, middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "form:token",
		Validator: func(key string, c echo.Context) (bool, error) {
			return key == config.GetApiToken(), nil
		},
	}))
	timeout := 12 * time.Hour
	e.PUT("/update/:id", mh.handleUpdate, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
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
	}))

	e.GET("/", mh.handleIndex)

	e.Any("/*", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/")
	})
}
