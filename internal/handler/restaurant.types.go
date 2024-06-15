package handler

import "gitlab.unjx.de/flohoss/mittag/internal/config"

type Restaurant struct {
	ID       string             `json:"id"`
	Icon     string             `json:"icon"`
	Name     string             `json:"name"`
	PageURL  string             `json:"page_url"`
	Address  string             `json:"address"`
	RestDays []config.DayOfWeek `json:"rest_days"`
	Phone    string             `json:"phone"`
	Group    config.Group       `json:"group"`
	Menu     config.Menu        `json:"menu"`
}

func ReduceRestaurant(restaurant *config.Restaurant) Restaurant {
	return Restaurant{
		ID:       restaurant.ID,
		Icon:     restaurant.Icon,
		Name:     restaurant.Name,
		PageURL:  restaurant.PageURL,
		Address:  restaurant.Address,
		RestDays: restaurant.RestDays,
		Phone:    restaurant.Phone,
		Group:    restaurant.Group,
		Menu:     restaurant.Menu,
	}
}
