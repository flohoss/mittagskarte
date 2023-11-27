package fetch

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func DownloadHtml(url string, http_one bool) (*goquery.Document, error) {
	slog.Debug("downloading html", "url", url)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Custom Agent")
	client := http.DefaultClient
	if http_one {
		slog.Debug("requesting with HTTP/1.1", "url", url)
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
