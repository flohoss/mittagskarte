package controller

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/date"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var headings = []string{"MONTAG", "MITTWOCH", "DONNERSTAG", "FREITAG", "ALTERNATIVESSEN", "ODER"}

func containsHeading(s string) bool {
	for _, heading := range headings {
		if strings.Contains(s, heading) {
			return true
		}
	}
	return false
}

func (c *Controller) handleFass(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}
	_, currentWeekNumber := time.Now().ISOWeek()

	box := doc.Find(fmt.Sprintf("div.sqrpara:contains('KW %d')", currentWeekNumber)).First()

	dateRegex := regexp.MustCompile(`(\d{2})\s?.?\s?\S+\s?(\d{2})\s?.?\s?(\S+)\s?(\d{4})`)
	groups := dateRegex.FindStringSubmatch(box.Text())
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
		box.Each(func(i int, s *goquery.Selection) {
			s.Find("tr").Each(func(i int, s *goquery.Selection) {
				if containsHeading(s.Text()) {
					f := Food{CardID: card.ID}
					s.Find("td").Each(func(i int, s *goquery.Selection) {
						switch i {
						case 0:
							f.Day = caser.String(strings.TrimSpace(s.Text()))
						case 1:
							f.Name = strings.TrimSpace(s.Text())
						}
					})
					f.Price = 10.00
					food = append(food, f)
				}
			})
		})

		c.orm.CreateInBatches(&food, len(food))
	}
}
