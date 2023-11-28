package mittag

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"code.sajari.com/docconv"
	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"gitlab.unjx.de/flohoss/mittag/internal/helper"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

type LiveInformation struct {
	HTMLPages    []*goquery.Document
	DownloadUrl  string
	FileLocation string
	RawText      string
}

func (l *LiveInformation) fetchAndStoreHtmlPage(url string, httpOne bool) error {
	page, err := fetch.DownloadHtml(url, httpOne)
	if err != nil {
		slog.Error("could not download html page", "url", url, "err", err)
		return err
	}
	l.HTMLPages = append(l.HTMLPages, page)
	l.RawText += "\n" + page.Text()
	return nil
}

func (l *LiveInformation) fetchAndStoreFile(id string, url string, httpOne bool, existingFileHash string) (string, error) {
	file, err := fetch.DownloadFile(id, url, httpOne)
	if err != nil {
		slog.Error("could not download file", "url", url, "err", err)
		return existingFileHash, err
	}

	hash, err := helper.GenerateHash(file)
	if err != nil {
		return existingFileHash, err
	}

	if hash == existingFileHash {
		errMsg := "file is identical, will not parse"
		slog.Debug(errMsg, "path", file)
		return existingFileHash, &os.PathError{Op: "stat", Path: file, Err: errors.New(errMsg)}
	}

	helper.RemoveAllOtherFiles(file)

	l.FileLocation = file
	return hash, nil
}

func (l *LiveInformation) findUrlInPage(r *Retrieve) error {
	downloadUrl := ""
	if r.Regex != "" {
		replaced := helper.ReplacePlaceholder(r.Regex)
		expr := regexp.MustCompile("(?i)" + replaced)
		res := expr.FindStringSubmatch(l.RawText)
		if len(res) > 1 {
			downloadUrl = res[1]
		}
	} else {
		present := false
		downloadUrl, present = l.HTMLPages[len(l.HTMLPages)-1].Find(helper.ReplacePlaceholder(r.JQuery)).First().Attr(r.Attribute)
		if !present {
			errMsg := "could not find url with given information"
			slog.Error(errMsg, "jquery", r.JQuery, "attribute", r.Attribute)
			return errors.New(errMsg)
		}
	}
	l.DownloadUrl = r.Prefix + downloadUrl
	return nil
}

func (l *LiveInformation) parseAndStoreFileText(c *Configuration) error {
	if len(c.Download.Cropping) != 0 {
		pngPath, err := convert.ConvertPdfToPng(l.FileLocation, false)
		if err != nil {
			return err
		}
		for i, crop := range c.Download.Cropping {
			res, err := convert.CropMenu(pngPath, fmt.Sprintf("%s-%d", c.Restaurant.ID, i), crop.Crop, crop.Gravity)
			if err != nil {
				continue
			}
			ocr, err := docconv.ConvertPath(res)
			if err != nil {
				slog.Error("could not parse file", "file", pngPath)
				continue
			}
			if crop.Keep {
				os.Rename(res, pngPath)
			} else {
				os.Remove(res)
			}
			l.RawText += "\n" + ocr.Body
			l.FileLocation = pngPath
		}
	} else {
		ocr, err := docconv.ConvertPath(l.FileLocation)
		if err != nil {
			slog.Error("could not parse file", "file", l.FileLocation)
			return err
		}
		l.RawText += "\n" + ocr.Body
	}
	return nil
}

func (l *LiveInformation) prepareFileForPublic(id string) error {
	delete := false
	if strings.Contains(l.FileLocation, "converted") {
		delete = true
	}
	webpFile, err := convert.ConvertToWebp(l.FileLocation, id, delete)
	if err != nil {
		return err
	}
	file := strings.Split(webpFile, "/")
	newFile := PublicLocation + file[len(file)-1]
	os.Rename(webpFile, newFile)
	slog.Debug("file moved", "old", webpFile, "new", newFile)
	l.FileLocation = newFile
	return nil
}
