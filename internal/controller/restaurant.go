package controller

import (
	"log/slog"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
	"gorm.io/gorm"
)

type RestaurantData struct {
	Title      string
	Navigation [][]restaurant.Restaurant
	Restaurant restaurant.Restaurant
}

func (c *Controller) RenderRestaurants(ctx echo.Context) error {
	var restaurant restaurant.Restaurant
	found := c.orm.Where("id = ?", ctx.Param("id")).Preload("Card.Food", func(db *gorm.DB) *gorm.DB {
		return db.Order("id")
	}).Find(&restaurant).RowsAffected
	if found == 0 {
		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	}
	return ctx.Render(http.StatusOK, "restaurants", RestaurantData{Title: restaurant.Name, Navigation: c.Navigation, Restaurant: restaurant})
}

func (c *Controller) UpdateRestaurants(ctx echo.Context) error {
	id := ctx.QueryParam("id")
	var result restaurant.Restaurant
	if id == "random" {
		c.setRandomRestaurant()
	} else if id != "" {
		affected := c.orm.Where("id = ?", id).Preload("Card").Find(&result).RowsAffected
		if affected == 0 {
			return ctx.NoContent(http.StatusNotFound)
		}
		go func() {
			card, err := result.Update()
			if err != nil {
				slog.Error("cannot update restaurant", "err", err)
			} else {
				c.orm.Create(&card)
			}
		}()
	} else {
		go c.UpdateAllRestaurants()
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) UpdateAllRestaurants() {
	restaurants := restaurant.GetRestaurants(c.orm)
	var cards []restaurant.Card
	for _, r := range restaurants {
		card, err := r.Update()
		if err != nil {
			slog.Error("cannot update all restaurants", "err", err)
		} else {
			cards = append(cards, card)
		}
	}
	c.orm.Where("1 = 1").Delete(&restaurant.Card{})
	c.orm.Create(&cards)
}

func getRandomRestaurantIndex(amount int) int {
	min := 0
	return min + rand.Intn(amount-min+1)
}

func (c *Controller) setRandomRestaurant() {
	var result []restaurant.Restaurant
	c.orm.Find(&result).Update("selected", false)
	for i, _ := range restaurant.Groups {
		amount := len(c.Navigation[i])
		if amount > 0 {
			random := c.Navigation[i][getRandomRestaurantIndex(amount-1)]
			c.orm.Model(&restaurant.Restaurant{}).Where("id = ?", random.ID).Update("selected", true)
			c.Navigation = restaurant.GetNavigation(c.orm)
		}
	}
}
