package fetch

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
)

func DownloadHtml(url string, httpVersion config.HTTPVersion) (*goquery.Document, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Custom Agent")
	client := http.DefaultClient

	slog.Debug("requesting html page", "url", url, "httpVersion", httpVersion)
	if httpVersion == config.HTTP1_0 || httpVersion == config.HTTP1_1 {
		client = &http.Client{
			Transport: &http.Transport{
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			},
		}
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("no 200 html status, status: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return doc, err

	}
	return doc, nil
}
