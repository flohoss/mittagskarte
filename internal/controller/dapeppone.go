package controller

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"gitlab.unjx.de/flohoss/mittag/internal/convert"
)

func (c *Controller) handleDaPeppone(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}
	fileLocation, err := restaurant.downloadCard(doc, "a:contains('Mittagskarte')", "href", "")
	if err != nil {
		return
	}
	imgUrl, err := convert.ConvertPdfToPng(fileLocation, "300")
	if err != nil {
		return
	}
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
		c.createCard(&card, []byte(doc.Find("a:contains('Mittagskarte')").Text()), `\d{2}\s?.\s?\S+(?:\s?.?\s?\d{4})?`, "02.01.2006")

		var food []Food
		wholeRegex := regexp.MustCompile(`Gericht\s\d\.?\d?:?\s(.+)\n€\s?(\d{1,2}\,\d{1,2})`)
		res := wholeRegex.FindAllSubmatch(text, -1)
		for _, g := range res {
			price, _ := strconv.ParseFloat(strings.Replace(string(g[2]), ",", ".", 1), 64)
			food = append(food, Food{
				CardID: card.ID,
				Name:   string(g[1]),
				Price:  price,
			})
		}
		c.orm.CreateInBatches(&food, len(food))
	}
	os.Remove(imgUrl)
}
