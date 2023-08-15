package fetch

import (
	"errors"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func DownloadHtml(url string) (*goquery.Document, error) {
	doc := &goquery.Document{}
	res, err := http.Get(url)
	if err != nil {
		return doc, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return doc, errors.New("no 200 html status")
	}
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return doc, err

	}
	return doc, nil
}
