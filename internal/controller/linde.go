package controller

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"gitlab.unjx.de/flohoss/mittag/internal/date"
)

func (c *Controller) handleLinde(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}
	month := date.GetMonthFromNumber(int(time.Now().Month()))
	fileLocation, err := restaurant.downloadCard(doc, fmt.Sprintf("td:contains('%s') > a", month), "href", "https://www.gasthauslinde.de")
	if err != nil {
		return
	}
	imgUrl, err := convert.ConvertPdfToPng(fileLocation, "300")
	if err != nil {
		return
	}
	convert.CropPng(imgUrl, "2300x1950+140+900", true)
	if restaurant.Card.ImageURL != convert.ReplaceEndingToWebp(imgUrl) {
		card := Card{
			RestaurantID: restaurant.ID,
			ImageURL:     convert.CreateWebp(imgUrl),
		}
		text, err := convert.OCR(card.ImageURL, 6)
		if err != nil {
			return
		}

		c.deleteCard(&restaurant.Card)
		weekNr, year, weekDay, start, begin, end := date.LindeRelevantDates("02.01.06")

		card.Description = fmt.Sprintf("%s: %s bis %s", month, begin, end)
		card.Year = year
		card.Week = weekNr
		card.Day = weekDay

		c.orm.Create(&card)
		var food []Food
		regex := regexp.MustCompile(`\w+\s?\|?\w+.?\s?\w{3}\s?\|?(.*)\n`)
		replaceRegex := regexp.MustCompile(`\||_|\d`)
		excludeRegex := regexp.MustCompile(`Sonntag\,|Samstag\,|Montag\,`)
		res := regex.FindAll(text, -1)
		for i, r := range res {
			f := regex.FindSubmatch(r)
			day := monday.Format(start.AddDate(0, 0, i), "Monday, 02.01.", monday.LocaleDeDE)
			if !excludeRegex.MatchString(day) {
				food = append(food, Food{
					CardID: card.ID,
					Name:   strings.TrimSpace(string(replaceRegex.ReplaceAllLiteral(f[1], []byte(" ")))),
					Day:    day,
					Price:  9.00,
				})
			}
		}
		c.orm.CreateInBatches(&food, len(food))
	}
	os.Remove(imgUrl)
}
