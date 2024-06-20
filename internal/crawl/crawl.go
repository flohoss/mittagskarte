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
	isFile      bool
	FinalUrl    string
	DocStorage  []*goquery.Document
}

func NewCrawler(initialUrl string, httpVersion config.HTTPVersion, navigate []config.Selector, isFile bool) *Crawler {
	c := &Crawler{
		httpVersion: httpVersion,
		navigate:    navigate,
		FinalUrl:    initialUrl,
		isFile:      isFile,
	}
	c.Crawl()
	return c
}

func (c *Crawler) Crawl() {
	if err := c.downloadHtml(-1); err != nil {
		slog.Error("could not download html", "url", c.FinalUrl, "err", err)
		return
	}
	for i := 0; i < len(c.navigate); i++ {
		c.FinalUrl = c.searchFinalUrl(i)
		if !c.isFile {
			if err := c.downloadHtml(i); err != nil {
				continue
			}
		}
	}
}

func (c *Crawler) downloadHtml(round int) error {
	doc, err := fetch.DownloadHtml(c.FinalUrl, c.httpVersion)
	if err != nil {
		return err
	}
	if round < 0 {
		c.DocStorage = []*goquery.Document{doc}
	} else {
		c.DocStorage = append(c.DocStorage, doc)
	}
	return nil
}

func (c *Crawler) searchFinalUrl(index int) string {
	var url string
	selector := c.navigate[index]
	if selector.Regex != "" {
		slog.Debug("searching for final url", "regex", selector.Regex)
		url = selector.RegexResult(c.DocStorage[len(c.DocStorage)-1].Text(), nil)
	} else if c.navigate[index].JQuery != "" {
		slog.Debug("searching for final url", "jquery", selector.JQuery, "attribute", selector.Attribute)
		url = selector.JQueryResult(c.DocStorage[len(c.DocStorage)-1])
	}
	if url == "" {
		slog.Warn("could not find final url")
		return c.FinalUrl
	}
	url = selector.Prefix + url
	slog.Debug("found url", "url", url)
	return url
}
