package controller

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (c *Controller) handleRatsstuben(restaurant *Restaurant) {
	doc, err := restaurant.downloadHtml()
	if err != nil {
		return
	}
	fileLocation, err := restaurant.downloadCard(doc, "div > p > a", "href", "")
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
		text, err := convert.OCR(card.ImageURL, 3)
		if err != nil {
			return
		}
		c.deleteCard(&restaurant.Card)
		c.createCard(&card, text, `\d{2}\s?.\s?\d{2}\s?.\s?\d{4}`, "02.01.2006")

		var food []Food
		prices := []float64{0, 0}
		lookingFor := []string{"TAGESESSEN", "VEGETARISCHES TAGESESSEN"}
		for i := 0; i < len(lookingFor); i++ {
			regex := regexp.MustCompile(fmt.Sprintf(`%s\s?(\d{1,3},\d{2})\s?€{1}`, lookingFor[i]))
			price, _ := strconv.ParseFloat(strings.Replace(string(regex.FindSubmatch(text)[1]), ",", ".", 1), 64)
			prices[i] = price
		}

		splits := []string{"MONTAG", "DIENSTAG", "MITTWOCH", "DONNERSTAG", "FREITAG", "BEILAGENSALAT"}
		caser := cases.Title(language.German)
		for i := 0; i < len(splits)-1; i++ {
			regex := regexp.MustCompile(fmt.Sprintf("(%s){1}([^.]+)(%s){1}", splits[i], splits[i+1]))
			replace := regexp.MustCompile(`(\n){1,2}`)
			foodRes := replace.ReplaceAll(regex.FindSubmatch(text)[2], []byte{' '})
			day := caser.Bytes(regex.FindSubmatch(text)[1])
			f := strings.Split(string(foodRes), " oder ")
			for index, singleFood := range f {
				food = append(food, Food{
					CardID: card.ID,
					Name:   singleFood,
					Price:  prices[index%2],
					Day:    string(day),
				})
			}
		}
		c.orm.CreateInBatches(&food, len(food))
	}
	os.Remove(imgUrl)
}
