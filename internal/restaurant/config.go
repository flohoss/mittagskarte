package restaurant

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"code.sajari.com/docconv"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
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

func (c *Configuration) getFinalDownloadUrl(downloadUrl string) (string, error) {
	if len(c.RetrieveDownloadUrl) > 0 {
		for _, d := range c.RetrieveDownloadUrl {
			slog.Debug("navigating to page", "page", downloadUrl)
			var err error
			doc, err := fetch.DownloadHtml(downloadUrl, c.HTTPOne)
			if err != nil {
				return "", err
			}
			var present bool
			downloadUrl, present = doc.Find(replacePlaceholder(d.JQuery)).First().Attr(d.Attribute)
			if !present {
				return "", errors.New("cannot navigate")
			}
			downloadUrl = d.Prefix + downloadUrl
		}
		slog.Debug("found final url", "url", downloadUrl)
		return downloadUrl, nil
	}
	return downloadUrl, nil
}

func (c *Configuration) downloadAndParseMenu(id string, downloadUrl string) ([]string, string, error) {
	var content []string
	filePath, err := fetch.DownloadFile(id, downloadUrl, c.HTTPOne)
	if err != nil {
		return content, "", err
	}

	filePath, err = convert.ConvertPdfToPng(filePath)
	if err != nil {
		return content, "", err
	}

	if len(c.Download.Cropping) != 0 {
		for i, c := range c.Download.Cropping {
			res, err := convert.CropMenu(filePath, fmt.Sprintf("%s-%d", id, i), c.Crop, c.Gravity)
			if err != nil {
				continue
			}
			ocr, err := docconv.ConvertPath(res)
			if err != nil {
				continue
			}
			if c.Keep {
				os.Rename(res, filePath)
			} else {
				os.Remove(res)
			}
			saveContentAsFile(id, fmt.Sprintf("-%d", i), ocr.Body)
			content = append(content, ocr.Body)
		}
	} else {
		ocr, err := docconv.ConvertPath(filePath)
		if err != nil {
			return content, "", err
		}
		saveContentAsFile(id, "", ocr.Body)
		content = append(content, ocr.Body)
	}
	filePath, err = convert.ConvertToWebp(filePath, id)
	if err != nil {
		return content, "", err
	}
	return content, filePath, nil
}
