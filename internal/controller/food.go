package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/goodsign/monday"
	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
)

type FoodData struct {
	Title       string
	Group       string
	Day         string
	Navigation  [][]restaurant.Restaurant
	Restaurants []restaurant.Restaurant
}

func (c *Controller) RenderFood(ctx echo.Context) error {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 8)
	if err != nil {
		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	}
	if id < int64(restaurant.Groups[0]) || id > int64(restaurant.Groups[len(restaurant.Groups)-1]) {
		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	}
	group := restaurant.Group(id)
	var restaurants []restaurant.Restaurant
	today := monday.Format(time.Now(), "Monday", monday.LocaleDeDE)
	options := []string{today, today + " (Vegetarisch)"}
	c.orm.Where(&restaurant.Restaurant{Group: group}).Preload("Card.Food", "Day IN ?", options).Order("Name").Find(&restaurants)
	return ctx.Render(http.StatusOK, "foods", FoodData{Title: "Mittag - " + today + " - " + group.String(), Group: group.String(), Day: today, Navigation: c.Navigation, Restaurants: restaurants})
}
