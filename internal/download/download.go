package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	maxRetries     = 5
	initialBackoff = 2 * time.Second
)

func File(downloadPath string, fullUrl string) (string, error) {
	// Parse the URL to get the file extension
	fileURL, err := url.Parse(fullUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}
	ext := filepath.Ext(fileURL.Path)
	if ext == "" {
		return "", fmt.Errorf("failed to determine file extension from URL %s", fullUrl)
	}
	// remove query parameter in case of resize or crop server side
	fileURL.RawQuery = ""

	// Create the file
	file, err := os.Create(downloadPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", downloadPath, err)
	}
	defer file.Close()

	// Set up exponential backoff for retry logic
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Make the HTTP GET request
		res, err := http.Get(fileURL.String())
		if err != nil {
			return "", fmt.Errorf("failed to fetch URL: %w", err)
		}
		defer res.Body.Close()

		// Check if the response status is 200 (OK)
		if res.StatusCode == http.StatusOK {
			// Copy the response body to the file
			_, err = io.Copy(file, res.Body)
			if err != nil {
				return "", fmt.Errorf("failed to write to file %s: %w", downloadPath, err)
			}
			// Successful download, return the file name
			return downloadPath, nil
		}

		// If we get a 503 Service Unavailable, apply retry logic
		if res.StatusCode == http.StatusServiceUnavailable || res.StatusCode == http.StatusTooManyRequests {
			// Check if the server provided a "Retry-After" header
			retryAfter := res.Header.Get("Retry-After")
			if retryAfter != "" {
				if waitTime, err := time.ParseDuration(retryAfter + "s"); err == nil {
					backoff = waitTime
				}
			}

			// Wait for the backoff period before retrying
			time.Sleep(backoff)

			// Exponentially increase the backoff time
			backoff *= 2
			continue
		}

		// For other non-200 status codes, return an error
		return "", fmt.Errorf("received non-200 status code: %d", res.StatusCode)
	}

	// If the max number of retries is reached, return an error
	return "", errors.New("max retries reached. Unable to download file")
}
