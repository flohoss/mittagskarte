package restaurant

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"code.sajari.com/docconv"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const ConfigLocation = "configs/restaurants/"

func parseConfig(path string) (Configuration, error) {
	slog.Debug("parsing config", "path", path)
	var config Configuration
	content, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}
	slog.Debug("config successfully parsed", "path", path)
	return config, nil
}

func parseAllConfigs() ([]Configuration, error) {
	var configurations []Configuration
	err := filepath.WalkDir(ConfigLocation, func(path string, info os.DirEntry, err error) error {
		if info.Type().IsRegular() {
			config, err := parseConfig(path)
			if err != nil {
				return err
			}
			configurations = append(configurations, config)
		}
		return nil
	})
	return configurations, err
}

func (c *Configuration) getFirstHtmlPage() error {
	doc, err := fetch.DownloadHtml(c.Restaurant.PageURL, c.HTTPOne)
	if err != nil {
		return err
	}
	c.htmlPages = append(c.htmlPages, doc)
	return nil
}

func (c *Configuration) getFinalHtmlPage() error {
	c.downloadUrl = c.Restaurant.PageURL
	if len(c.RetrieveDownloadUrl) > 0 {
		for _, d := range c.RetrieveDownloadUrl {
			slog.Debug("navigating to page", "page", c.downloadUrl)
			downloadUrl, present := c.htmlPages[len(c.htmlPages)-1].Find(replacePlaceholder(d.JQuery)).First().Attr(d.Attribute)
			if !present {
				return errors.New("cannot navigate")
			}
			c.downloadUrl = d.Prefix + downloadUrl
			doc, err := fetch.DownloadHtml(c.downloadUrl, c.HTTPOne)
			if err != nil {
				return err
			}
			c.htmlPages = append(c.htmlPages, doc)
		}
		return nil
	}
	return nil
}

func (c *Configuration) saveContentAsFile(suffix string, content string) error {
	folder := fetch.DownloadLocation + c.Restaurant.ID
	os.MkdirAll(folder, os.ModePerm)
	err := os.WriteFile(fmt.Sprintf("%s/%s%s.txt", folder, c.Restaurant.ID, suffix), []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (c *Configuration) downloadAndParseMenu() error {
	filePath, err := fetch.DownloadFile(c.Restaurant.ID, c.downloadUrl, c.HTTPOne)
	if err != nil {
		return err
	}

	filePath, err = convert.ConvertPdfToPng(filePath)
	if err != nil {
		return err
	}

	if len(c.Download.Cropping) != 0 {
		for i, crop := range c.Download.Cropping {
			res, err := convert.CropMenu(filePath, fmt.Sprintf("%s-%d", c.Restaurant.ID, i), crop.Crop, crop.Gravity)
			if err != nil {
				continue
			}
			ocr, err := docconv.ConvertPath(res)
			if err != nil {
				continue
			}
			if crop.Keep {
				os.Rename(res, filePath)
			} else {
				os.Remove(res)
			}
			c.content = append(c.content, ocr.Body)
			c.saveContentAsFile(fmt.Sprintf("-%d", i), ocr.Body)
		}
	} else {
		ocr, err := docconv.ConvertPath(filePath)
		if err != nil {
			return err
		}
		c.content = append(c.content, ocr.Body)
		c.saveContentAsFile("", ocr.Body)
	}
	filePath, err = convert.ConvertToWebp(filePath, c.Restaurant.ID)
	if err != nil {
		return err
	}
	c.card.ImageURL = filePath
	return nil
}

func (c *Configuration) parseDescription() error {
	if c.Menu.Description.Regex != "" {
		replaced := replacePlaceholder(c.Menu.Description.Regex)
		slog.Debug("description from regex", "regex", replaced)
		descriptionExpr := regexp.MustCompile("(?i)" + replaced)
		c.card.Description = descriptionExpr.FindString(c.content[0])
	} else if c.Menu.Description.JQuery != "" {
		replaced := replacePlaceholder(c.Menu.Description.JQuery)
		slog.Debug("description from jquery", "jquery", replaced)
		if c.Menu.Description.Attribute == "" {
			c.card.Description = c.htmlPages[0].Find(replaced).First().Text()
		} else {
			present := false
			c.card.Description, present = c.htmlPages[0].Find(replaced).First().Attr(c.Menu.Description.Attribute)
			if !present {
				return errors.New("cannot find jquery")
			}
		}
	} else if c.Menu.Description.Fixed != "" {
		slog.Debug("description fixed", "fixed", c.Menu.Description.Fixed)
		c.card.Description = c.Menu.Description.Fixed
	}
	caser := cases.Title(language.German)
	c.card.Description = caser.String(c.card.Description)
	return nil
}
