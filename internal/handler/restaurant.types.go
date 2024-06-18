package handler

import "gitlab.unjx.de/flohoss/mittag/internal/config"

type Restaurant struct {
	ID          string             `json:"id"`
	Price       uint8              `json:"price"`
	Icon        string             `json:"icon"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	PageURL     string             `json:"page_url"`
	Address     string             `json:"address"`
	RestDays    []config.DayOfWeek `json:"rest_days"`
	Phone       string             `json:"phone"`
	Group       config.Group       `json:"group"`
	Menu        config.Menu        `json:"menu"`
}

func ReduceRestaurant(restaurant *config.Restaurant) Restaurant {
	return Restaurant{
		ID:          restaurant.ID,
		Price:       restaurant.Price,
		Icon:        restaurant.Icon,
		Name:        restaurant.Name,
		Description: restaurant.Description,
		PageURL:     restaurant.PageURL,
		Address:     restaurant.Address,
		RestDays:    restaurant.RestDays,
		Phone:       restaurant.Phone,
		Group:       restaurant.Group,
		Menu:        restaurant.Menu,
	}
}
