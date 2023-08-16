package fetch

import (
	"errors"
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

func DownloadFile(id string, fullUrl string) (string, error) {
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
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(fullUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("response status is not 200")
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return fileName, nil
}
