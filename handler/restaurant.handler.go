package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.unjx.de/flohoss/mittag/services"
	"golang.org/x/time/rate"
)

type MittagHandler struct {
	mittag *services.Mittag
}

func NewMittagHandler(mittag *services.Mittag) *MittagHandler {
	return &MittagHandler{
		mittag: mittag,
	}
}

func (h *MittagHandler) GetAllRestaurantsOperation() huma.Operation {
	return huma.Operation{
		OperationID: "get-all-restaurants",
		Method:      http.MethodGet,
		Path:        "/restaurants",
		Summary:     "Get all restaurants",
		Description: "Retrieve a list of all restaurants with their details.",
		Tags:        []string{"Restaurants"},
	}
}

type Restaurants struct {
	Body map[string]*services.CleanRestaurant
}

func (h *MittagHandler) GetAllRestaurants(ctx context.Context, input *struct{}) (*Restaurants, error) {
	return &Restaurants{Body: h.mittag.GetAllRestaurants()}, nil
}

func (h *MittagHandler) GetRestaurantOperation() huma.Operation {
	return huma.Operation{
		OperationID: "get-restaurant",
		Method:      http.MethodGet,
		Path:        "/restaurants/{id}",
		Summary:     "Get a single restaurant",
		Description: "Retrieve details for a specific restaurant.",
		Tags:        []string{"Restaurants"},
	}
}

type Restaurant struct {
	Body *services.CleanRestaurant
}

func (h *MittagHandler) GetRestaurant(ctx context.Context, input *struct {
	ID string `path:"id" maxLength:"255" doc:"restaurant id" required:"true"`
}) (*Restaurant, error) {
	r, err := h.mittag.GetRestaurant(input.ID)
	if err != nil {
		return nil, huma.Error404NotFound(err.Error())
	}
	return &Restaurant{Body: r}, nil
}

func (h *MittagHandler) UploadMenuOperation() huma.Operation {
	return huma.Operation{
		OperationID: "upload-menu",
		Method:      http.MethodPost,
		Path:        "/restaurants/{id}",
		Summary:     "Upload a menu",
		Description: "Upload a menu for a specific restaurant.",
		Tags:        []string{"Restaurants"},
	}
}

func (h *MittagHandler) UploadMenu(ctx context.Context, input *struct {
	ID      string `path:"id" maxLength:"255" doc:"restaurant id" required:"true"`
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"file" required:"true" doc:"The text file to upload"`
	}]
}) (*struct{}, error) {
	formData := input.RawBody.Data()
	if !formData.File.IsSet {
		return nil, huma.Error400BadRequest("no file uploaded")
	}
	data, err := io.ReadAll(formData.File)
	if err != nil {
		return nil, huma.Error400BadRequest("failed to read uploaded file: " + err.Error())
	}
	_, err = h.mittag.UploadMenu(input.ID, data, formData.File.Filename)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	return nil, nil
}

func getRateLimiter() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(1.0 / (12.0 * float64(time.Hour))), // Allow 1 request per 12 hours
				Burst:     1,
				ExpiresIn: 12 * time.Hour,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			restaurant := ctx.Param("id")
			if restaurant == "" {
				restaurant = "none"
			}
			return restaurant, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusForbidden, "Error extracting identifier for rate limiting.")
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Zu viele Anfragen für dieses Restaurant (max 1/12h). Bitte versuchen Sie es in ein paar Minuten erneut.")
		},
	})
}

func (h *MittagHandler) RefreshRestaurantOperation() huma.Operation {
	return huma.Operation{
		OperationID: "refresh-restaurant",
		Method:      http.MethodPut,
		Path:        "/restaurants/{id}",
		Summary:     "Refresh a menu",
		Description: "Refresh a menu for a specific restaurant.",
		Tags:        []string{"Restaurants"},
	}
}

func (h *MittagHandler) RefreshRestaurant(ctx context.Context, input *struct {
	ID string `path:"id" maxLength:"255" doc:"restaurant id" required:"true"`
}) (*struct{}, error) {
	_, err := h.mittag.UpdateRestaurant(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	return nil, nil
}
