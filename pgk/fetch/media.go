package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadFile(filePath string, url string) error {
	// Make an HTTP GET request to the image URL
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("failed to download file: status code %d", response.StatusCode)
	}
	defer response.Body.Close()

	// Create a file to save the image
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the image data to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
