package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/services"
)

type Restaurant struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Tags      []string    `json:"tags"`
	PageUrl   string      `json:"url"`
	Address   string      `json:"address"`
	RestDays  []string    `json:"rest_days"`
	Phone     string      `json:"phone"`
	Group     string      `json:"group"`
	CreatedAt config.Date `json:"created_at"`
	Menu      config.Menu `json:"menu"`
}

type RestaurantHandler struct {
	mittag *services.Mittag
}

func NewRestaurantHandler(mittag *services.Mittag) *RestaurantHandler {
	return &RestaurantHandler{
		mittag: mittag,
	}
}

func (jh *RestaurantHandler) listRestaurants() []Restaurant {
	retaurants := config.GetRestaurants()
	restaurants := make([]Restaurant, 0, len(retaurants))
	for _, r := range retaurants {
		restaurants = append(restaurants, Restaurant{
			ID:        r.ID,
			Name:      r.Name,
			Tags:      r.Tags,
			PageUrl:   r.PageUrl,
			Address:   r.Address,
			RestDays:  r.RestDays,
			Phone:     r.Phone,
			Group:     r.Group,
			CreatedAt: r.CreatedAt,
			Menu:      r.Menu,
		})
	}
	return restaurants
}

func (jh *RestaurantHandler) listRestaurantsOperation() huma.Operation {
	return huma.Operation{
		OperationID: "get-restaurants",
		Method:      http.MethodGet,
		Path:        "/api/restaurants",
		Summary:     "Get restaurants with their current menu and details.",
		Description: "Get restaurants with their current menu and details.",
		Tags:        []string{"Restaurants"},
	}
}

type Restaurants struct {
	Body []Restaurant `json:"body"`
}

func (jh *RestaurantHandler) listRestaurantsHandler(ctx context.Context, input *struct{}) (*Restaurants, error) {
	restaurants := jh.listRestaurants()
	return &Restaurants{Body: restaurants}, nil
}

func (jh *RestaurantHandler) getRestaurantOperation() huma.Operation {
	return huma.Operation{
		OperationID: "get-restaurant",
		Method:      http.MethodGet,
		Path:        "/api/restaurants/{id}",
		Summary:     "Get a restaurant by ID.",
		Description: "Get a restaurant by ID.",
		Tags:        []string{"Restaurants"},
	}
}

type RestaurantResponse struct {
	Body Restaurant `json:"body"`
}

func (jh *RestaurantHandler) getRestaurantHandler(ctx context.Context, input *struct {
	ID string `path:"id"`
}) (*RestaurantResponse, error) {
	restaurant, err := config.GetRestaurant(input.ID)
	if err != nil {
		return nil, huma.Error404NotFound(err.Error())
	}
	return &RestaurantResponse{Body: Restaurant{
		ID:        restaurant.ID,
		Name:      restaurant.Name,
		Tags:      restaurant.Tags,
		PageUrl:   restaurant.PageUrl,
		Address:   restaurant.Address,
		RestDays:  restaurant.RestDays,
		Phone:     restaurant.Phone,
		Group:     restaurant.Group,
		CreatedAt: restaurant.CreatedAt,
		Menu:      restaurant.Menu,
	}}, nil
}
