package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"
)

const DownloadLocation = "static/downloads/"

func init() {
	os.MkdirAll(DownloadLocation, os.ModePerm)
}

func DownloadFile(id string, fullUrl string) (string, error) {
	fileURL, err := url.Parse(fullUrl)
	if err != nil {
		zap.S().Warnf("Could not parse url %s", fullUrl)
		return "", err
	}
	path := fileURL.Path
	folder := DownloadLocation + id
	os.MkdirAll(folder, os.ModePerm)
	segments := strings.Split(path, "/")
	fileName := fmt.Sprintf("%s/%s", folder, segments[len(segments)-1])

	file, err := os.Create(fileName)
	if err != nil {
		zap.S().Warnf("Could not create file %s", file)
		return "", err
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(fullUrl)
	if err != nil {
		zap.S().Warnf("Could not download file %s", file)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errorText := "Response status is not 200"
		zap.L().Warn(errorText)
		return "", fmt.Errorf(strings.ToLower(errorText))
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		zap.S().Warnf("Could not copy file %s", file)
		return "", err
	}
	defer file.Close()
	return fileName, nil
}
