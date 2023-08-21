package restaurant

import (
	"errors"
	"log/slog"
	"os"
	"regexp"

	"code.sajari.com/docconv"
	"github.com/PuerkitoBio/goquery"
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

	card.Description, err = parseDescription(&config, &content, doc)
	if err != nil {
		return card, err
	}

	var allFood []Food
	for _, f := range config.Menu.Food {
		allFood = append(allFood, Food{
			Name:        f.getName(&content),
			Day:         f.getDay(&content),
			Price:       f.getPrice(&content),
			Description: f.getDescription(&content),
		})
	}
	card.Food = allFood

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
		}
		slog.Debug("found final url", "url", downloadUrl)
		return downloadUrl, doc, nil
	}
	return downloadUrl, nil, nil
}

func downloadFile(id string, config *Configuration, downloadUrl string) (string, string, error) {
	imageURL, err := fetch.DownloadFile(id, config.Download.Prefix+downloadUrl)
	if err != nil {
		return "", "", err
	}
	slog.Debug("scanning file", "path", imageURL)
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

func parseDescription(config *Configuration, content *string, doc *goquery.Document) (string, error) {
	description := ""
	if config.Menu.Description.Regex != "" {
		replaced := replacePlaceholder(config.Menu.Description.Regex)
		slog.Debug("description from regex", "regex", replaced)
		descriptionExpr := regexp.MustCompile(replaced)
		description = descriptionExpr.FindString(*content)
	} else if config.Menu.Description.JQuery != "" {
		slog.Debug("description from jquery", "jquery", config.Menu.Description.JQuery)
		if config.Menu.Description.Attribute == "" {
			description = doc.Find(config.Menu.Description.JQuery).First().Text()
		} else {
			present := false
			description, present = doc.Find(config.Menu.Description.JQuery).First().Attr(config.Menu.Description.Attribute)
			if !present {
				return "", errors.New("cannot find jquery")
			}
		}
	} else if config.Menu.Description.Fixed != "" {
		slog.Debug("description fixed", "fixed", config.Menu.Description.Fixed)
		return config.Menu.Description.Fixed, nil
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
