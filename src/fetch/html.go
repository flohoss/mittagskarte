package fetch

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func GetHtmlOfPage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err

	}
	return doc, nil
}
