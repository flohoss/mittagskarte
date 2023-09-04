package restaurant

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"code.sajari.com/docconv"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/otiai10/gosseract/v2"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
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

	downloadUrl, doc, err := getFinalDownloadUrl(&config, r.PageURL)
	if err != nil {
		return card, err
	}

	var content []string
	if config.Download.IsFile {
		content, card.ImageURL, err = downloadAndParseMenu(r.ID, &config, downloadUrl)
	} else {
		content, doc, err = downloadHtml(r.ID, downloadUrl)
	}
	if err != nil {
		return card, err
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

func getFinalDownloadUrl(config *Configuration, downloadUrl string) (string, *goquery.Document, error) {
	if len(config.RetrieveDownloadUrl) > 0 {
		doc := &goquery.Document{}
		for _, d := range config.RetrieveDownloadUrl {
			slog.Debug("navigating to page", "page", downloadUrl)
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
			downloadUrl = d.Prefix + downloadUrl
		}
		slog.Debug("found final url", "url", downloadUrl)
		return downloadUrl, doc, nil
	}
	return downloadUrl, nil, nil
}

func downloadAndParseMenu(id string, config *Configuration, downloadUrl string) ([]string, string, error) {
	var content []string
	imageURL, err := fetch.DownloadFile(id, config.Download.Prefix+downloadUrl)
	if err != nil {
		return content, "", err
	}
	slog.Debug("scanning file", "path", imageURL)
	if len(config.Download.Cropping) != 0 {
		for i, c := range config.Download.Cropping {
			res, err := convert.CropMenu(imageURL, fmt.Sprintf("%s-%d", id, i), c.Crop, c.Gravity)
			if err != nil {
				continue
			}
			ocr, err := docconv.ConvertPath(res)
			if err != nil {
				continue
			}
			os.Remove(res)
			saveContentAsFile(id, fmt.Sprintf("-%d", i), ocr.Body)
			content = append(content, ocr.Body)
		}
	} else {
		ocr, err := docconv.ConvertPath(imageURL)
		if err != nil {
			return content, "", err
		}
		saveContentAsFile(id, "", ocr.Body)
		content = append(content, ocr.Body)
	}
	imageURL, err = convert.ConvertToWebp(imageURL, id, config.Download.TrimEdges)
	if err != nil {
		return content, "", err
	}
	return content, imageURL, nil
}

func downloadHtml(id string, downloadUrl string) ([]string, *goquery.Document, error) {
	doc, err := fetch.DownloadHtml(downloadUrl)
	if err != nil {
		return []string{}, doc, err
	}
	saveContentAsFile(id, "", doc.Text())
	return []string{doc.Text()}, doc, nil
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
