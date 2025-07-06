package handler

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
)

func longCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000")
		return next(c)
	}
}

func BearerTokenMiddleware(api huma.API, expectedToken string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
			return
		}

		providedToken := parts[1]

		if providedToken != expectedToken {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Invalid bearer token")
			return
		}

		next(ctx)
	}
}

func SetupRouter(e *echo.Echo, handler *MittagHandler, token string) {
	config := huma.DefaultConfig("Mittag API", "1.0.0")
	config.OpenAPIPath = "/api/openapi"
	config.SchemasPath = "/api/schemas"
	h := humaecho.New(e, config)

	e.GET("/api/docs", func(ctx echo.Context) error {
		return ctx.HTML(http.StatusOK, `<!doctype html>
			<html>
				<head>
					<title>API Reference</title>
					<meta charset="utf-8" />
					<meta name="viewport" content="width=device-width, initial-scale=1" />
				</head>
				<body>
					<script id="api-reference" data-url="/api/openapi.json"></script>
					<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
				</body>
			</html>`,
		)
	})
	e.Renderer = initTemplates()

	huma.Register(h, handler.GetAllRestaurantsOperation(), handler.GetAllRestaurants)
	huma.Register(h, handler.GetRestaurantOperation(), handler.GetRestaurant)
	huma.Register(h, handler.RefreshRestaurantOperation(h, token), handler.RefreshRestaurant)
	huma.Register(h, handler.UploadMenuOperation(h, token), handler.UploadMenu)

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	public := e.Group("/data", longCacheLifetime)
	public.Static("/thumbnails", "data/thumbnails")

	assets := e.Group("/assets", longCacheLifetime)
	assets.Static("/", "web/assets")

	favicon := e.Group("/favicon", longCacheLifetime)
	favicon.Static("/", "web/favicon")

	storage := e.Group("/storage", longCacheLifetime)
	storage.Static("/downloads", "storage/downloads")

	e.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "index.html", nil)
	})
}
