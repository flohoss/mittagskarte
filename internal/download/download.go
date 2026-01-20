package download

import (
	"fmt"
	"os/exec"
)

func Curl(downloadPath, fullURL string) (string, error) {
	cmd := exec.Command("curl", "-L", "-o", downloadPath, fullURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("curl failed: %v, output: %s", err, string(output))
	}
	return downloadPath, nil
}
