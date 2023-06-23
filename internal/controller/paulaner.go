package controller

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"gitlab.unjx.de/flohoss/mittag/internal/convert"
)

func (c *Controller) handlePaulaner(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}
	restaurant.PageURL, _ = doc.Find("a:contains('Zum Mittagsangebot')").First().Attr("href")
	doc, err = restaurant.downloadHtml()
	if err != nil {
		return
	}
	fileLocation, err := restaurant.downloadCard(doc, "article > div > embed", "data", "")
	if err != nil {
		return
	}
	imgUrl, err := convert.ConvertPdfToPng(fileLocation, "300")
	if err != nil {
		return
	}
	convert.CropPng(imgUrl, "1820x2150+400+600", true)

	if restaurant.Card.ImageURL != convert.ReplaceEndingToWebp(imgUrl) {
		card := Card{
			RestaurantID: restaurant.ID,
			ImageURL:     convert.CreateWebp(imgUrl),
		}
		text, err := convert.OCR(card.ImageURL, 3)
		if err != nil {
			return
		}
		c.deleteCard(&restaurant.Card)
		err = c.createCard(&card, text, `\d{2}\s?.\s?\S+(?:\s?.?\s?\d{4})?`, "02.January2006")
		if err != nil {
			return
		}

		var food []Food
		wholeRegex := regexp.MustCompile(`[^\d]+\n€{1}\s?\d{1,3},\d{2}`)
		priceRegex := regexp.MustCompile(`€{1}\s?\d{1,3},\d{2}`)
		res := wholeRegex.FindAll(text, -1)
		for _, single := range res {
			priceStr := string(priceRegex.Find(single))
			price, _ := strconv.ParseFloat(strings.Replace(restaurant.removeEuro(priceStr), ",", ".", 1), 64)
			food = append(food, Food{
				CardID: card.ID,
				Name:   strings.Replace(strings.TrimSpace(strings.Replace(string(single), priceStr, "", 1)), "\n", " ", -1),
				Price:  price,
			})
		}
		c.orm.CreateInBatches(&food, len(food))
	}
	os.Remove(imgUrl)
}
