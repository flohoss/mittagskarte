package restaurant

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/otiai10/gosseract/v2"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func GetNavigation(orm *gorm.DB) [][]Restaurant {
	var navigation [][]Restaurant
	for _, g := range Groups {
		var restaurants []Restaurant
		orm.Where(&Restaurant{Group: g}).Select("ID", "Name", "Selected", "Latitude", "Longitude", "Group").Order("Name").Find(&restaurants)
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
	slog.Debug("updating restaurant", "name", r.Name)
	card := Card{RestaurantID: r.ID}
	config, err := parseConfig(ConfigLocation + r.ID + ".json")
	if err != nil {
		return card, err
	}

	downloadUrl, err := config.getFinalDownloadUrl(r.PageURL)
	if err != nil {
		return card, err
	}

	doc, err := fetch.DownloadHtml(downloadUrl, config.HTTPOne)
	if err != nil {
		return card, err
	}
	saveContentAsFile(r.ID, "", doc.Text())
	content := []string{doc.Text()}

	if config.Download.IsFile {
		content, card.ImageURL, err = config.downloadAndParseMenu(r.ID, downloadUrl)
		if err != nil {
			return card, err
		}
	}

	if len(content) > 0 {
		card.Description, err = parseDescription(&config, &content[0], doc)
		if err != nil {
			return card, err
		}
	}

	for _, c := range content {
		card.Food = append(card.Food, config.getAllFood(&c, doc)...)
	}
	return card, nil
}

func parseDescription(config *Configuration, content *string, doc *goquery.Document) (string, error) {
	description := ""
	if config.Menu.Description.Regex != "" {
		replaced := replacePlaceholder(config.Menu.Description.Regex)
		slog.Debug("description from regex", "regex", replaced)
		descriptionExpr := regexp.MustCompile("(?i)" + replaced)
		description = descriptionExpr.FindString(*content)
	} else if config.Menu.Description.JQuery != "" {
		replaced := replacePlaceholder(config.Menu.Description.JQuery)
		slog.Debug("description from jquery", "jquery", replaced)
		if config.Menu.Description.Attribute == "" {
			description = doc.Find(replaced).First().Text()
		} else {
			present := false
			description, present = doc.Find(replaced).First().Attr(config.Menu.Description.Attribute)
			if !present {
				return "", errors.New("cannot find jquery")
			}
		}
	} else if config.Menu.Description.Fixed != "" {
		slog.Debug("description fixed", "fixed", config.Menu.Description.Fixed)
		description = config.Menu.Description.Fixed
	}
	caser := cases.Title(language.German)
	return caser.String(description), nil
}

func saveContentAsFile(id string, suffix string, content string) error {
	folder := fetch.DownloadLocation + id
	os.MkdirAll(folder, os.ModePerm)
	err := os.WriteFile(fmt.Sprintf("%s/%s%s.txt", folder, id, suffix), []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
