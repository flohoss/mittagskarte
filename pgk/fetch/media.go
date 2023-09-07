package fetch

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const DownloadLocation = "storage/downloads/"

func init() {
	os.MkdirAll(DownloadLocation, os.ModePerm)
}

func DownloadFile(id string, fullUrl string, http_one bool) (string, error) {
	fileURL, err := url.Parse(fullUrl)
	if err != nil {
		return "", err
	}
	path := fileURL.Path
	folder := DownloadLocation + id
	os.MkdirAll(folder, os.ModePerm)
	segments := strings.Split(path, "/")
	fileName := fmt.Sprintf("%s/%s", folder, segments[len(segments)-1])

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("GET", fullUrl, nil)
	req.Header.Set("User-Agent", "Custom Agent")
	client := http.DefaultClient
	if http_one {
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
