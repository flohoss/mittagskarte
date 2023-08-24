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

func ConvertToWebp(fileLocation string, resultName string, trim bool) (string, error) {
	dir := filepath.Dir(fileLocation)
	result := fmt.Sprintf("%s/%s.webp", dir, resultName)
	app := "convert"
	args := []string{}
	if trim {
		args = []string{"-trim"}
	}
	args = append(args, []string{"-strip", "-density", "300", "-alpha", "Remove", fileLocation, "-quality", "90", result}...)
	slog.Debug("converting to webp", "path", fileLocation, "command", args)
	out, err := exec.Command(app, args...).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	os.Remove(fileLocation)
	slog.Debug("file successfully converted", "path", result)
	return result, nil
}

func CropMenu(fileLocation string, resultName string, cropping string, gravity string) (string, error) {
	var err error
	dir := filepath.Dir(fileLocation)
	ext := filepath.Ext(fileLocation)
	if ext == ".pdf" {
		fileLocation, err = ConvertPdfToPng(fileLocation)
		if err != nil {
			return "", err
		}
		ext = ".png"
	}
	result := fmt.Sprintf("%s/%s%s", dir, resultName, ext)
	app := "convert"
	args := []string{}
	if gravity != "" {
		args = []string{"-gravity", gravity}
	}
	args = append(args, []string{"-crop", cropping, fileLocation, result}...)
	slog.Debug("cropping menu", "path", fileLocation, "command", args)
	out, err := exec.Command(app, args...).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	if ext == ".png" {
		os.Remove(fileLocation)
	}
	slog.Debug("file successfully cropped", "path", result)
	return result, nil
}

func ConvertPdfToPng(fileLocation string) (string, error) {
	result := strings.Replace(fileLocation, ".pdf", "", 1)
	slog.Debug("converting pdf to png", "path", fileLocation)
	out, err := exec.Command("pdftoppm", "-singlefile", "-r", "300", "-png", fileLocation, result).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	return result + ".png", nil
}
