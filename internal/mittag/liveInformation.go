package mittag

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"gitlab.unjx.de/flohoss/mittag/internal/helper"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

type LiveInformation struct {
	HTMLPages          []*goquery.Document
	FileDownloadUrl    string
	StoredFileLocation string
	FileText           string
}

func (l *LiveInformation) fetchAndStoreHtmlPage(url string, httpOne bool) error {
	page, err := fetch.DownloadHtml(url, httpOne)
	if err != nil {
		slog.Error("could not download html page", "url", url, "err", err)
		return err
	}
	l.HTMLPages = append(l.HTMLPages, page)
	return nil
}

func (l *LiveInformation) fetchAndStoreFile(id string, url string, httpOne bool) error {
	fileName := fetch.GetFilename(id, url)
	if _, err := os.Stat(fileName); err == nil {
		errMsg := "file exists already"
		slog.Warn(errMsg, "path", fileName)
		return &os.PathError{Op: "stat", Path: fileName, Err: errors.New(errMsg)}
	}

	file, err := fetch.DownloadFile(id, url, httpOne)
	if err != nil {
		slog.Error("could not download file", "url", url, "err", err)
		return err
	}
	l.StoredFileLocation = file
	return nil
}

func (l *LiveInformation) findDownloadUrlInPage(r *Retrieve) error {
	downloadUrl, present := l.HTMLPages[len(l.HTMLPages)-1].Find(helper.ReplacePlaceholder(r.JQuery)).First().Attr(r.Attribute)
	if !present {
		errMsg := "could not find url with given information"
		slog.Error(errMsg, "jquery", r.JQuery, "attribute", r.Attribute)
		return errors.New(errMsg)
	}
	l.FileDownloadUrl = r.Prefix + downloadUrl
	return nil
}

func (l *LiveInformation) parseAndStoreFileText() error {
	text, err := helper.ParseMenu(l.StoredFileLocation)
	if err != nil {
		slog.Error("could not parse file", "file", l.StoredFileLocation)
		return err
	}
	l.FileText = text
	return nil
}

func (l *LiveInformation) prepareFileForPublic(id string) error {
	webpFile, err := convert.ConvertToWebp(l.StoredFileLocation, id, false)
	if err != nil {
		return err
	}
	file := strings.Split(webpFile, "/")
	newFile := PublicLocation + file[len(file)-1]
	os.Rename(webpFile, newFile)
	slog.Debug("file moved", "old", webpFile, "new", newFile)
	l.StoredFileLocation = newFile
	return nil
}
