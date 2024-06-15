package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
)

type RestaurantHandler struct {
	restaurants map[string]*config.Restaurant
}

func NewHandler(restaurants map[string]*config.Restaurant) *RestaurantHandler {
	return &RestaurantHandler{
		restaurants: restaurants,
	}
}

// GetAllRestaurants
//
//	@Produce	json
//	@Tags		restaurants
//	@Success	200	{object}	map[string]Restaurant	"ok"
//	@Router		/restaurants [get]
func (h *RestaurantHandler) GetAllRestaurants(ctx echo.Context) error {
	restaurants := make(map[string]Restaurant)
	for _, restaurant := range h.restaurants {
		restaurants[restaurant.ID] = ReduceRestaurant(restaurant)
	}
	return ctx.JSON(http.StatusOK, restaurants)
}

// GetAllRestaurantsGrouped
//
//	@Produce	json
//	@Tags		groups
//	@Success	200	{object}	map[config.Group][]Restaurant	"ok"
//	@Router		/groups [get]
func (h *RestaurantHandler) GetAllRestaurantsGrouped(ctx echo.Context) error {
	groups := make(map[config.Group][]Restaurant)
	for _, group := range config.AllGroups {
		groups[group] = []Restaurant{}
	}
	for _, restaurant := range h.restaurants {
		groups[restaurant.Group] = append(groups[restaurant.Group], ReduceRestaurant(restaurant))
	}
	return ctx.JSON(http.StatusOK, groups)
}

// GetRestaurant
//
//	@Produce	json
//	@Tags		restaurants
//	@Param		id	path		string			true	"Restaurant ID"
//	@Success	200	{object}	Restaurant		"ok"
//	@Failure	404	{object}	echo.HTTPError	"Can not find ID"
//	@Router		/restaurants/{id} [get]
func (h *RestaurantHandler) GetRestaurant(ctx echo.Context) error {
	id := ctx.Param("id")
	restaurant, ok := h.restaurants[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "Can not find ID")
	}
	return ctx.JSON(http.StatusOK, ReduceRestaurant(restaurant))
}
