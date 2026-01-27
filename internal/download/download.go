package download

import (
	"fmt"
	"os/exec"

	"github.com/getsentry/sentry-go"
)

func Curl(downloadPath, fullURL string) (string, error) {
	cmd := exec.Command("curl", "-L", "-o", downloadPath, fullURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sentry.CaptureException(err)
		sentry.CaptureException(err)
		return "", fmt.Errorf("curl failed: %s", string(output))
	}
	return downloadPath, nil
}
