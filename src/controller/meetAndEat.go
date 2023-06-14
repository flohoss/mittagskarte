package controller

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"mittag/convert"
)

func (c *Controller) handleMeetAndEat(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}
	imgUrl, err := restaurant.downloadCard(doc, "div#Speisekarte > a > span > img", "src", "")
	if err != nil {
		return
	}

	if restaurant.Card.ImageURL != convert.ReplaceEndingToWebp(imgUrl) {
		card := Card{
			RestaurantID: restaurant.ID,
			ImageURL:     convert.CreateWebp(imgUrl),
		}
		text, err := convert.OCR(card.ImageURL, 11)
		if err != nil {
			return
		}
		c.deleteCard(&restaurant.Card)
		c.createCard(&card, text, `\d{2}\s?.\s?\d{2}\s?.\s?\d{2}`, "02.01.06")

		var food []Food
		wholeRegex := regexp.MustCompile(`(.*)\s*€{1}\s?(\d{1,3},\d{2})`)
		res := wholeRegex.FindAll(text, -1)
		for _, single := range res {
			groups := wholeRegex.FindSubmatch(single)
			if strings.Contains(string(groups[1]), "+") {
				continue
			}
			price, _ := strconv.ParseFloat(strings.Replace(string(groups[2]), ",", ".", 1), 64)
			food = append(food, Food{
				CardID: card.ID,
				Name:   string(groups[1]),
				Price:  price,
			})
		}
		c.orm.CreateInBatches(&food, len(food))
	}
	os.Remove(imgUrl)
}
