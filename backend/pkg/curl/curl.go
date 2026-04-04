package curl

import (
	"fmt"
	"os/exec"
)

func Download(downloadPath, fullURL string) (string, error) {
	cmd := exec.Command("curl", "-L", "-o", downloadPath, fullURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("curl failed: %s", string(output))
	}

	return downloadPath, nil
}
