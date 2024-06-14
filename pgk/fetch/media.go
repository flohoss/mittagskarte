package fetch

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gitlab.unjx.de/flohoss/mittag/internal/config"
)

const DownloadLocation = "storage/downloads/"

func init() {
	os.MkdirAll(DownloadLocation, os.ModePerm)
}

func ParseFileNameFromUrl(id string, fullUrl string) string {
	fileURL, _ := url.Parse(fullUrl)
	path := fileURL.Path
	folder := DownloadLocation + id
	os.MkdirAll(folder, os.ModePerm)
	segments := strings.Split(path, "/")
	return fmt.Sprintf("%s/%s", folder, segments[len(segments)-1])
}

func DownloadFile(id string, fullUrl string, httpVersion config.HTTPVersion) (string, error) {
	slog.Debug("downloading file", "url", fullUrl)

	fileName := ParseFileNameFromUrl(id, fullUrl)
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

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
	defer file.Close()
	return fileName, nil
}
