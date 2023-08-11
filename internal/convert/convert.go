package convert

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func ReplaceEndingToWebp(fileLocation string) string {
	ext := filepath.Ext(fileLocation)
	newFile := strings.Replace(fileLocation, ext, ".webp", 1)
	return newFile
}

func ConvertPdfToWebp(fileLocation string, resultName string, dpi string, trim bool) (string, error) {
	dir := filepath.Dir(fileLocation)
	result := fmt.Sprintf("%s/%s.webp", dir, resultName)
	app := "convert"
	args := []string{"-strip", "-density", dpi, "-alpha", "Remove", fileLocation, "-quality", "90", result}
	if trim {
		args = append(args, "-trim")
	}
	zap.L().Debug("converting pdf to webp", zap.Strings("command", args))
	out, err := exec.Command(app, args...).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	os.Remove(fileLocation)
	return result, nil
}
