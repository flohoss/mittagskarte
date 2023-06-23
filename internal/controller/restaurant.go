package controller

import (
	"math/rand"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/fetch"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var WhiteSpaceRegex = regexp.MustCompile(`\s`)

type RestaurantData struct {
	Title      string
	Default    Default
	Restaurant Restaurant
}

func (c *Controller) RenderRestaurants(ctx echo.Context) error {
	var restaurant Restaurant
	found := c.orm.Where("id = ?", ctx.Param("id")).Preload("Card", func(db *gorm.DB) *gorm.DB {
		return db.Order("year").Order("week").Order("day")
	}).Preload("Card.Food", func(db *gorm.DB) *gorm.DB {
		return db.Order("id")
	}).Find(&restaurant).RowsAffected
	if found == 0 {
		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	}
	return ctx.Render(http.StatusOK, "restaurants", RestaurantData{Title: restaurant.Name, Default: c.Default, Restaurant: restaurant})
}

func (c *Controller) UpdateRestaurants(ctx echo.Context) error {
	id := ctx.QueryParam("id")
	var result Restaurant
	if id != "" {
		affected := c.orm.Where("id = ?", id).Preload("Card").Find(&result).RowsAffected
		if affected == 0 {
			return ctx.NoContent(http.StatusNotFound)
		}
		go c.updateRestaurantData(&result)
	} else {
		c.setRandomRestaurant()
		go c.updateData()
	}
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) restaurants() []Restaurant {
	var result []Restaurant
	c.orm.Find(&result)
	return result
}

func (c *Controller) restaurantsJoinCardJoinFood() []Restaurant {
	var result []Restaurant
	c.orm.Model(&Restaurant{}).Preload("Card", func(db *gorm.DB) *gorm.DB {
		return db.Order("year").Order("week").Order("day")
	}).Preload("Card.Food", func(db *gorm.DB) *gorm.DB {
		return db.Order("id")
	}).Order("name").Find(&result)
	return result
}

func (c *Controller) getRandomRestaurantIndex(amount int) int {
	min := 0
	return min + rand.Intn(amount-min)
}

func (c *Controller) setRandomRestaurant() {
	var result []Restaurant
	c.orm.Find(&result).Update("selected", false)
	amount := len(c.Default.StuttgartRestaurants)
	if amount > 0 {
		random := c.Default.StuttgartRestaurants[c.getRandomRestaurantIndex(amount-1)]
		c.orm.Model(&Restaurant{}).Where("id = ?", random.ID).Update("selected", true)
		c.setupDefaults()
	}
}

func (c *Controller) updateRestaurantData(restaurant *Restaurant) {
	zap.L().Debug("Updating restaurant", zap.String("name", restaurant.Name))
	switch restaurant.Name {
	case "Meet & Eat":
		c.handleMeetAndEat(restaurant)
	case "Paulaner":
		c.handlePaulaner(restaurant)
	case "Picchiorosso":
		c.handlePicchiorosso(restaurant)
	case "SW34":
		c.handleSW34(restaurant)
	case "Schwedenscheuer":
		c.handleSchwedenscheuer(restaurant)
	case "Ratsstuben":
		c.handleRatsstuben(restaurant)
	case "Fass":
		c.handleFass(restaurant)
	case "Linde":
		c.handleLinde(restaurant)
	case "Da Peppone":
		c.handleDaPeppone(restaurant)
	}
}

func (r *Restaurant) downloadHtml() (*goquery.Document, error) {
	res, err := http.Get(r.PageURL)
	if err != nil {
		zap.S().Errorf("Failed GET request of %s -> %s", r.PageURL, err.Error())
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		zap.L().Error("Request did not result in status 200")
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		zap.L().Error("Could not parse document from response body")
		return nil, err
	}
	return doc, nil
}

func (r *Restaurant) downloadCard(doc *goquery.Document, jquery string, attr string, urlPrefix string) (string, error) {
	downloadUrl, _ := doc.Find(jquery).First().Attr(attr)
	fileLocation, err := fetch.DownloadFile(r.ID, urlPrefix+downloadUrl)
	if err != nil {
		zap.L().Error("Cannot download card", zap.Error(err))
		return "", err
	}
	return fileLocation, nil
}

func (r *Restaurant) removeEuro(s string) string {
	return strings.TrimSpace(strings.Replace(s, "€", "", 1))
}
