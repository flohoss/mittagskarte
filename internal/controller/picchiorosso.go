package controller

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/goodsign/monday"
	"gitlab.unjx.de/flohoss/mittag/internal/date"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (c *Controller) handlePicchiorosso(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}

	heading := strings.TrimSpace(doc.Find("h2.av-special-heading-tag:contains('Tagesgerichte')").Text())
	dateRegex := regexp.MustCompile(`(\d{2})\s?.\s?(\d{2})\s?.\s?(\d{4})`)
	groups := dateRegex.FindStringSubmatch(heading)
	if len(groups) < 4 {
		zap.L().Debug("No header detected", zap.String("groups", fmt.Sprintf("%s", groups)))
		return
	}
	today := fmt.Sprintf("%s.%s.%s", groups[1], groups[2], groups[3])
	weekNr, year, weekDay, dateStr, goDate := date.ExtractRelevantFromDateString(today, "02.01.2006")
	card := Card{
		RestaurantID: restaurant.ID,
		Description:  fmt.Sprintf("%s: %s", monday.Format(goDate, "Monday", monday.LocaleDeDE), dateStr),
		Year:         year,
		Week:         weekNr,
		Day:          weekDay,
	}

	if restaurant.Card.Year != card.Year || restaurant.Card.Week != card.Week || restaurant.Card.Day != card.Day {
		c.deleteCard(&restaurant.Card)
		c.orm.Create(&card)

		var food []Food
		caser := cases.Title(language.Italian)
		doc.Find("ul.av-catalogue-list").Children().Each(func(i int, s *goquery.Selection) {
			price, _ := strconv.ParseFloat(restaurant.removeEuro(s.Find(".av-catalogue-price").Text()), 64)
			food = append(food, Food{
				CardID:      card.ID,
				Name:        strings.TrimSpace(caser.String(s.Find(".av-catalogue-title").Text())),
				Price:       price,
				Description: strings.TrimSpace(s.Find(".av-catalogue-content").Text()),
			})
		})
		c.orm.CreateInBatches(&food, len(food))
	}
}
