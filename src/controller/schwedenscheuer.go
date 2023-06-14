package controller

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"mittag/date"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

func (c *Controller) handleSchwedenscheuer(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}

	heading := strings.TrimSpace(doc.Find("table").Text())
	dateRegex := regexp.MustCompile(`(\d{2})\s?.?\s?(\d{2})\s?.?\s?(\d{4})?\s?\S+\s?(\d{2})\s?.?\s?(\d{2})\s?.?\s?(\d{4})?`)
	groups := dateRegex.FindStringSubmatch(heading)
	if len(groups) < 6 {
		c.log.Debug("No header detected", zap.String("groups", fmt.Sprintf("%s", groups)))
		return
	}
	fromDate := fmt.Sprintf("%s.%s.%s", groups[1], groups[2], groups[6])
	toDate := fmt.Sprintf("%s.%s.%s", groups[4], groups[5], groups[6])
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
		doc.Find("table > tbody > tr").Each(func(i int, s *goquery.Selection) {
			aligned, _ := s.Find("td").First().Attr("align")
			if aligned == "left" && i <= 7 {
				f := Food{CardID: card.ID}
				s.Find("td").Each(func(i int, s *goquery.Selection) {
					switch i {
					case 0:
						f.Day = strings.TrimSpace(strings.Replace(s.Text(), ":", "", 1))
					case 1:
						f.Name = strings.TrimSpace(strings.Replace(s.Text(), ":", "", 1))
					case 2:
						price, _ := strconv.ParseFloat(strings.Replace(strings.TrimSpace(strings.Replace(s.Text(), ":", "", 1)), ",", ".", 1), 64)
						f.Price = price
					}
				})
				food = append(food, f)
			}
		})
		c.orm.CreateInBatches(&food, len(food))
	}
}
