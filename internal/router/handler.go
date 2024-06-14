package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
)

type Handler struct {
	restaurants map[string]*config.Restaurant
}

func NewHandler(restaurants map[string]*config.Restaurant) *Handler {
	return &Handler{
		restaurants: restaurants,
	}
}

func (h *Handler) GetAllRestaurants(ctx echo.Context) error {
	restaurants := make(map[string]*Restaurant)
	for _, restaurant := range h.restaurants {
		restaurants[restaurant.ID] = ReduceRestaurant(restaurant)
	}
	return ctx.JSON(http.StatusOK, restaurants)
}

func (h *Handler) GetRestaurant(ctx echo.Context) error {
	id := ctx.Param("id")
	restaurant, ok := h.restaurants[id]
	if !ok {
		return ctx.JSON(http.StatusNotFound, nil)
	}
	return ctx.JSON(http.StatusOK, ReduceRestaurant(restaurant))
}
