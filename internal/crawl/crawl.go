package crawl

import (
	"log/slog"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

type Crawler struct {
	httpVersion config.HTTPVersion
	navigate    []config.Selector
	FinalUrl    string
	DocStorage  []*goquery.Document
}

func NewCrawler(initialUrl string, httpVersion config.HTTPVersion, navigate []config.Selector) *Crawler {
	c := &Crawler{
		httpVersion: httpVersion,
		navigate:    navigate,
		FinalUrl:    initialUrl,
	}
	c.Crawl()
	return c
}

func (c *Crawler) Crawl() error {
	doc, err := fetch.DownloadHtml(c.FinalUrl, c.httpVersion)
	if err != nil {
		return err
	}
	slog.Debug("downloaded html", "url", c.FinalUrl)
	c.DocStorage = append(c.DocStorage, doc)
	for i := 0; i < len(c.navigate); i++ {
		c.FinalUrl = c.SearchFinalUrl(i)
		doc, err := fetch.DownloadHtml(c.FinalUrl, c.httpVersion)
		if err != nil {
			return err
		}
		c.DocStorage = append(c.DocStorage, doc)
	}
	return nil
}

func (c *Crawler) SearchFinalUrl(index int) string {
	var url string
	selector := c.navigate[index]
	if selector.Regex != "" {
		slog.Debug("searching for final url", "regex", selector.Regex)
		url = selector.RegexResult(c.DocStorage[len(c.DocStorage)-1].Text())
	} else if c.navigate[index].JQuery != "" {
		slog.Debug("searching for final url", "jquery", selector.JQuery, "attribute", selector.Attribute)
		url = selector.JQueryResult(c.DocStorage[len(c.DocStorage)-1])
	}
	slog.Debug("found final url", "url", url)
	return url
}
