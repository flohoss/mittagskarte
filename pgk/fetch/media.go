package fetch

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"gitlab.unjx.de/flohoss/mittag/internal/config"
)

const DownloadLocation = "storage/downloads/"

func init() {
	os.MkdirAll(DownloadLocation, os.ModePerm)
}

func DownloadFile(id string, fullUrl string, httpVersion config.HTTPVersion) (string, error) {
	fileURL, _ := url.Parse(fullUrl)
	ext := filepath.Ext(fileURL.Path)
	fileName := filepath.Join(DownloadLocation, id+ext)
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	req, _ := http.NewRequest("GET", fullUrl, nil)
	req.Header.Set("User-Agent", "Custom Agent")
	client := http.DefaultClient
	if httpVersion == config.HTTP1_0 || httpVersion == config.HTTP1_1 {
		client = &http.Client{
			Transport: &http.Transport{
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			},
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("no 200 html status, status: %d", resp.StatusCode)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	return fileName, nil
}
