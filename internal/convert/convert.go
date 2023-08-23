package convert

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	args := []string{}
	if trim {
		args = []string{"-trim"}
	}
	args = append(args, []string{"-strip", "-density", dpi, "-alpha", "Remove", fileLocation, "-quality", "90", result}...)
	slog.Debug("converting pdf to webp", "path", fileLocation, "command", args)
	out, err := exec.Command(app, args...).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	os.Remove(fileLocation)
	slog.Debug("file successfully converted", "path", result)
	return result, nil
}

func CutPdf(fileLocation string, resultName string, cropping string, gravity string) (string, error) {
	dir := filepath.Dir(fileLocation)
	ext := filepath.Ext(fileLocation)
	result := fmt.Sprintf("%s/%s%s", dir, resultName, ext)
	app := "convert"
	args := []string{}
	if gravity != "" {
		args = []string{"-gravity", gravity}
	}
	args = append(args, []string{"-crop", cropping, fileLocation, result}...)
	slog.Debug("cropping pdf", "path", fileLocation, "command", args)
	out, err := exec.Command(app, args...).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	slog.Debug("file successfully cropped", "path", result)
	return result, nil
}
