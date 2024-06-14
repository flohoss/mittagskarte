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

// GetAllRestaurants
//
//	@Produce	json
//	@Success	200	{object}	Restaurant	"ok"
//	@Router		/restaurants [get]
func (h *Handler) GetAllRestaurants(ctx echo.Context) error {
	restaurants := make(map[string]*Restaurant)
	for _, restaurant := range h.restaurants {
		restaurants[restaurant.ID] = ReduceRestaurant(restaurant)
	}
	return ctx.JSON(http.StatusOK, restaurants)
}

// GetRestaurant
//
//	@Produce	json
//	@Param		id	path		string			true	"Restaurant ID"
//	@Success	200	{object}	Restaurant		"ok"
//	@Failure	404	{object}	echo.HTTPError	"Can not find ID"
//	@Router		/restaurants/{id} [get]
func (h *Handler) GetRestaurant(ctx echo.Context) error {
	id := ctx.Param("id")
	restaurant, ok := h.restaurants[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "Can not find ID")
	}
	return ctx.JSON(http.StatusOK, ReduceRestaurant(restaurant))
}
