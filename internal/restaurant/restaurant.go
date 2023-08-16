package restaurant

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"code.sajari.com/docconv"
	"github.com/PuerkitoBio/goquery"
	"github.com/goodsign/monday"
	_ "github.com/otiai10/gosseract/v2"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
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

func posInArray(str string, arr []string) int {
	for i, s := range arr {
		if s == str {
			return i
		}
	}
	return -1
}

func (r *Restaurant) Update() (Card, error) {
	slog.Info("updating restaurant", "name", r.Name)
	card := Card{RestaurantID: r.ID}
	config, err := parseConfig(ConfigLocation + r.ID + ".json")
	if err != nil {
		return card, err
	}

	downloadUrl, doc, err := getFinalDownloadUrl(&config, r.PageURL)
	if err != nil {
		return card, err
	}

	var content string
	if config.Download.IsFile {
		content, card.ImageURL, err = downloadFile(r.ID, &config, downloadUrl)
	} else {
		content, doc, err = downloadHtml(downloadUrl)
	}
	if err != nil {
		return card, err
	}
	saveContentAsFile(r.ID, content)

	card.Description, err = parseDescription(&config, content, doc)
	if err != nil {
		return card, err
	}
	return card, nil
}

func replacePlaceholder(input string) string {
	if strings.Contains(input, "%KW%") {
		_, weekNr := time.Now().ISOWeek()
		return strings.Replace(input, "%KW%", fmt.Sprintf("%d", weekNr), -1)
	}
	if strings.Contains(input, "%month%") {
		return strings.Replace(input, "%month%", monday.Format(time.Now(), "January", monday.LocaleDeDE), -1)
	}
	return input
}

func getFinalDownloadUrl(config *Configuration, downloadUrl string) (string, *goquery.Document, error) {
	if len(config.RetrieveDownloadUrl) > 0 {
		doc := &goquery.Document{}
		for _, d := range config.RetrieveDownloadUrl {
			slog.Info("navigating to page", "page", downloadUrl)
			var err error
			doc, err = fetch.DownloadHtml(downloadUrl)
			if err != nil {
				return "", doc, err
			}
			var present bool
			downloadUrl, present = doc.Find(replacePlaceholder(d.JQuery)).First().Attr(d.Attribute)
			if !present {
				return "", doc, errors.New("cannot navigate")
			}
		}
		slog.Info("found final url", "url", downloadUrl)
		return downloadUrl, doc, nil
	}
	return downloadUrl, nil, nil
}

func downloadFile(id string, config *Configuration, downloadUrl string) (string, string, error) {
	imageURL, err := fetch.DownloadFile(id, config.Download.Prefix+downloadUrl)
	if err != nil {
		return "", "", err
	}
	slog.Info("scanning file", "path", imageURL)
	ocr, err := docconv.ConvertPath(imageURL)
	if err != nil {
		return "", "", err
	}
	imageURL, err = convert.ConvertPdfToWebp(imageURL, id, "300", config.Download.TrimEdges)
	if err != nil {
		return "", "", err
	}
	return ocr.Body, imageURL, nil
}

func downloadHtml(downloadUrl string) (string, *goquery.Document, error) {
	doc, err := fetch.DownloadHtml(downloadUrl)
	if err != nil {
		return "", doc, err
	}
	return doc.Text(), doc, nil
}

func parseDescription(config *Configuration, content string, doc *goquery.Document) (string, error) {
	description := ""
	if config.Menu.Description.Regex != "" {
		replaced := replacePlaceholder(config.Menu.Description.Regex)
		slog.Info("description from regex", "regex", replaced)
		descriptionExpr := regexp.MustCompile(replaced)
		description = descriptionExpr.FindString(content)
	} else if config.Menu.Description.JQuery != "" {
		slog.Info("description from jquery", "jquery", config.Menu.Description.JQuery)
		if config.Menu.Description.Attribute == "" {
			description = doc.Find(config.Menu.Description.JQuery).First().Text()
		} else {
			present := false
			description, present = doc.Find(config.Menu.Description.JQuery).First().Attr(config.Menu.Description.Attribute)
			if !present {
				return "", errors.New("cannot find jquery")
			}
		}
	}
	return description, nil
}

func saveContentAsFile(id string, content string) error {
	folder := fetch.DownloadLocation + id
	os.MkdirAll(folder, os.ModePerm)
	err := os.WriteFile(folder+"/text.txt", []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
