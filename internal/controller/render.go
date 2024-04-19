package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/mittag"
)

type TemplateData struct {
	BaseData       BaseData
	RestaurantData RestaurantData
	GroupData      GroupData
}

type BaseData struct {
	Title          string
	Configurations map[string]*mittag.Configuration
	Groups         []string
}

type RestaurantData struct {
	Restaurant mittag.Restaurant
	Card       mittag.Card
	Refreshed  string
	Updated    string
}

type GroupData struct {
	FilteredConfigurations map[string]*mittag.Configuration
	Cards                  []mittag.Card
	Today                  string
	Group                  string
}

func (c *Controller) RenderSettings(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "settings", TemplateData{
		BaseData: BaseData{Title: "Mittag - Einstellungen", Configurations: c.mittag.Configurations, Groups: mittag.Groups},
	})
}

func (c *Controller) RenderCountdown(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "countdown", TemplateData{
		BaseData: BaseData{Title: "Mittag - Countdown", Configurations: c.mittag.Configurations, Groups: mittag.Groups},
	})
}

func (c *Controller) RenderRestaurants(ctx echo.Context) error {
	exists, conf := c.mittag.DoesConfigurationExist(ctx.Param("id"))
	if !exists {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/")
	}
	var card mittag.Card
	c.mittag.GetORM().Where("restaurant_id = ?", conf.Restaurant.ID).Preload("Food").Preload("Map").Find(&card)
	return ctx.Render(http.StatusOK, "restaurants", TemplateData{
		BaseData:       BaseData{Title: "Mittag - " + conf.Restaurant.Name, Configurations: c.mittag.Configurations, Groups: mittag.Groups},
		RestaurantData: RestaurantData{Restaurant: conf.Restaurant, Card: card, Refreshed: c.humanizer.NaturalTime(card.Refreshed), Updated: c.humanizer.NaturalTime(card.UpdatedAt)},
	})
}
