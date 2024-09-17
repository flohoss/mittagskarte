package fetch

import (
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
	fileURL, err := url.Parse(fullUrl)
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(fileURL.Path)
	fileName := filepath.Join(DownloadLocation, id+ext)
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	res, err := http.Get(fullUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("no 200 html status, status: %d", res.StatusCode)
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return "", err
	}
	return fileName, nil
}
