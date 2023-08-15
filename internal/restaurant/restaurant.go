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

	content, img, doc, err := downloadFileOrHtml(r.ID, &config, downloadUrl)
	if err != nil {
		return card, err
	}
	card.ImageURL = img

	parseDescription(&config, content, doc)
	saveContentAsFile(r.ID, content)
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

func getFinalDownloadUrl(config *Configuration, downloadUrl string) (string, *goquery.Document, error) {
	doc := &goquery.Document{}
	if len(config.RetrieveDownloadUrl) > 0 {
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
	}
	slog.Info("found final url", "url", downloadUrl)
	return downloadUrl, doc, nil
}

func downloadFileOrHtml(id string, config *Configuration, downloadUrl string) (string, string, *goquery.Document, error) {
	content, img := "", ""
	doc := &goquery.Document{}
	var err error

	if config.Download.IsFile {
		content, img, err = downloadFile(id, config, downloadUrl)
	} else {
		content, doc, err = downloadHtml(downloadUrl)
	}
	return content, img, doc, err
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
	webpUrl, err := convert.ConvertPdfToWebp(imageURL, id, "300", config.Download.TrimEdges)
	if err != nil {
		return "", "", err
	}
	return ocr.Body, webpUrl, nil
}

func downloadHtml(downloadUrl string) (string, *goquery.Document, error) {
	slog.Info("downloading html", "url", downloadUrl)
	doc, err := fetch.DownloadHtml(downloadUrl)
	if err != nil {
		return "", &goquery.Document{}, err
	}
	return doc.Text(), doc, nil
}

func parseDescription(config *Configuration, content string, doc *goquery.Document) string {
	description := ""
	if strings.Compare(config.Menu.Description.Regex, "") != 0 {
		descriptionExpr := regexp.MustCompile(replacePlaceholder(config.Menu.Description.Regex))
		description = descriptionExpr.FindString(content)
	} else if strings.Compare(config.Menu.Description.JQuery, "") != 0 {
		if strings.Compare(config.Menu.Description.Attribute, "") == 0 {
			description = doc.Find(replacePlaceholder(config.Menu.Description.JQuery)).First().Text()
		} else {
			description, _ = doc.Find(replacePlaceholder(config.Menu.Description.JQuery)).First().Attr(config.Menu.Description.Attribute)
		}
	}
	return description
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
