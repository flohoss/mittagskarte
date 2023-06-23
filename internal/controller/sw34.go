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

func containsDay(s string) bool {
	for _, day := range monday.GetLongDays(monday.LocaleDeDE) {
		if strings.Contains(s, day) {
			return true
		}
	}
	return false
}

func (c *Controller) handleSW34(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}

	heading := strings.TrimSpace(doc.Find("h4.elementor-heading-title:contains('Wochenkarte')").First().Text())
	dateRegex := regexp.MustCompile(`(\d{2})\s?.?\s?\S+\s?(\d{2})\s?.?\s?(\S+)\s?(\d{4})`)
	groups := dateRegex.FindStringSubmatch(heading)
	if len(groups) < 5 {
		zap.L().Debug("No header detected", zap.String("groups", fmt.Sprintf("%s", groups)))
		return
	}
	fromDate := fmt.Sprintf("%s.%02d.%s", groups[1], date.GetMonthFromString(string(groups[3])), groups[4])
	toDate := fmt.Sprintf("%s.%02d.%s", groups[2], date.GetMonthFromString(string(groups[3])), groups[4])
	weekNr, year, weekDay, from, to := date.ExtractRelevantFromDateRange(fromDate, toDate, "02.01.2006")
	card := Card{
		RestaurantID: restaurant.ID,
		Description:  fmt.Sprintf("Woche %d: %s bis %s", weekNr, from, to),
		Year:         year,
		Week:         weekNr,
		Day:          weekDay,
	}

	if restaurant.Card.Year != card.Year || restaurant.Card.Week != card.Week {
		c.deleteCard(&restaurant.Card)
		c.orm.Create(&card)

		var food []Food
		caser := cases.Title(language.German)
		doc.Find("h4:contains('Tagesempfehlung')").First().ParentsFiltered("div.elementor-widget-wrap").First().Find("div.elementor-price-list-text").Each(func(i int, s *goquery.Selection) {
			day := caser.String(strings.TrimSpace(s.Find(".elementor-price-list-title").Text()))
			if containsDay(day) {
				priceStr := s.Find(".elementor-price-list-price").Text()
				price, _ := strconv.ParseFloat(strings.Replace(restaurant.removeEuro(priceStr), ",", ".", 1), 64)
				food = append(food, Food{
					CardID: card.ID,
					Name:   strings.TrimSpace(s.Find(".elementor-price-list-description").Text()),
					Day:    day,
					Price:  price,
				})
			}
		})
		c.orm.CreateInBatches(&food, len(food))
	}
}
