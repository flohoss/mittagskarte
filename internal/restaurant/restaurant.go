package restaurant

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"code.sajari.com/docconv"
	"github.com/PuerkitoBio/goquery"
	"github.com/goodsign/monday"
	_ "github.com/otiai10/gosseract/v2"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func UpdateAllRestaurants(orm *gorm.DB) {
	restaurants := GetRestaurants(orm)
	var cards []Card
	for _, r := range restaurants {
		card, err := r.Update()
		if err != nil {
			zap.L().Error(err.Error())
		} else {
			cards = append(cards, card)
		}
	}
	orm.Where("1 = 1").Delete(&Card{})
	orm.Create(&cards)
}

func GetNavigation(orm *gorm.DB) [][]Restaurant {
	var navigation [][]Restaurant
	for _, g := range Groups {
		var restaurants []Restaurant
		orm.Where(&Restaurant{Group: g}).Select("ID", "Name", "Selected", "Latitude", "Longitude").Order("Name").Find(&restaurants)
		navigation = append(navigation, restaurants)
	}
	return navigation
}

func GetRestaurants(orm *gorm.DB) []Restaurant {
	var result []Restaurant
	orm.Model(&Restaurant{}).Preload("Card.Food", func(db *gorm.DB) *gorm.DB {
		return db.Order("id")
	}).Order("name").Find(&result)
	return result
}

func (r *Restaurant) Update() (Card, error) {
	var card Card
	config, err := parseConfig(ConfigLocation + r.ID + ".json")
	if err != nil {
		return card, err
	}
	var doc *goquery.Document
	var content, fileLocation string
	var present bool
	downloadUrl := r.PageURL
	if len(config.Download) > 0 {
		for _, d := range config.Download {
			doc, err = fetch.DownloadHtml(downloadUrl)
			if err != nil {
				return card, err
			}
			downloadUrl, present = doc.Find(replacePlaceholder(d.JQuery)).First().Attr(d.Attribute)
			if !present {
				return card, errors.New("cannot find the menu of the restaurant")
			}
		}
		fileLocation, err = fetch.DownloadFile(r.ID, config.DownloadPrefix+downloadUrl)
		if err != nil {
			return card, err
		}
		ocr, err := docconv.ConvertPath(fileLocation)
		if err != nil {
			return card, err
		}
		fileLocation, err = convert.ConvertPdfToPng(fileLocation, r.ID, "300")
		if err != nil {
			return card, err
		}
		fileLocation, err = convert.CreateWebp(fileLocation)
		if err != nil {
			return card, err
		}
		content = ocr.Body
	} else {
		doc, err = fetch.DownloadHtml(downloadUrl)
		if err != nil {
			return card, err
		}
		if len(config.Redirect) > 0 {
			for _, r := range config.Redirect {
				downloadUrl, present = doc.Find(replacePlaceholder(r.JQuery)).First().Attr(r.Attribute)
				if !present {
					return card, errors.New("cannot find the redirect button")
				}
				doc, err = fetch.DownloadHtml(config.RedirectPrefix + downloadUrl)
				if err != nil {
					return card, err
				}
			}
			content = doc.Text()
		}
		content = doc.Text()
	}
	folder := fetch.DownloadLocation + r.ID
	os.MkdirAll(folder, os.ModePerm)
	err = os.WriteFile(folder+"/text.txt", []byte(content), os.ModePerm)
	if err != nil {
		return card, err
	}

	var descrResult string
	if config.DescriptionRegex != "" {
		config.DescriptionRegex = replacePlaceholder(config.DescriptionRegex)
		descrExpr := regexp.MustCompile(config.DescriptionRegex)
		if config.DescriptionInHtml {
			descrResult = descrExpr.FindString(doc.Text())
		} else {
			descrResult = descrExpr.FindString(content)
		}
	}

	var food []Food
	if config.FoodRegex != "" {
		config.FoodRegex = replacePlaceholder(config.FoodRegex)
		foodExpr := regexp.MustCompile(config.FoodRegex)
		foodResult := foodExpr.FindAllStringSubmatch(content, -1)
		for i, r := range foodResult {
			if config.MaxFood != 0 && i >= config.MaxFood {
				break
			}
			var f Food
			if config.Positions.Name > 0 {
				f.Name = strings.ReplaceAll(strings.TrimSpace(r[config.Positions.Name]), "\n", " ")
			}
			if config.Positions.Day > 0 {
				caser := cases.Title(language.German)
				f.Day = caser.String(r[config.Positions.Day])
			}
			if config.FixPrice != 0 {
				f.Price = config.FixPrice
			} else if config.Positions.Price > 0 {
				price, _ := strconv.ParseFloat(strings.Replace(r[config.Positions.Price], ",", ".", 1), 64)
				f.Price = price
			}
			if config.Positions.Description > 0 {
				f.Description = r[config.Positions.Description]
			}
			food = append(food, f)
		}
	}

	card = Card{
		RestaurantID: r.ID,
		Description:  descrResult,
		ImageURL:     fileLocation,
		Food:         food,
		CreatedAt:    0,
	}
	card.Description = strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, card.Description)
	card.Description = strings.Replace(strings.TrimSpace(card.Description), "�", "", 1)
	return card, nil
}

func replacePlaceholder(input string) string {
	if strings.Contains(input, "%KW%") {
		_, weekNr := time.Now().ISOWeek()
		return strings.Replace(input, "%KW%", fmt.Sprintf("%d", weekNr), 1)
	}
	if strings.Contains(input, "%month%") {
		return strings.Replace(input, "%month%", monday.Format(time.Now(), "January", monday.LocaleDeDE), 1)
	}
	return input
}
