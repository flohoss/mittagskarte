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

func ConvertToWebp(fileLocation string, resultName string) (string, error) {
	dir := filepath.Dir(fileLocation)
	result := fmt.Sprintf("%s/%s.webp", dir, resultName)
	out, err := exec.Command("convert", "-strip", "-density", "300", "-alpha", "Remove", fileLocation, "-quality", "90", result).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	os.Remove(fileLocation)
	slog.Debug("file successfully converted to webp", "path", result)
	return result, nil
}

func ConvertToPng(fileLocation string) (string, error) {
	ext := filepath.Ext(fileLocation)
	if ext != ".png" && ext != ".pdf" {
		result := strings.Replace(fileLocation, ext, ".png", 1)
		out, err := exec.Command("convert", fileLocation, result).CombinedOutput()
		if err != nil {
			return "", errors.New(string(out))
		}
		os.Rename(result, fileLocation)
		slog.Debug("file successfully converted to png", "path", result)
		return result, nil
	}
	return fileLocation, nil
}

func CropMenu(fileLocation string, resultName string, cropping string, gravity string) (string, error) {
	dir := filepath.Dir(fileLocation)
	ext := filepath.Ext(fileLocation)
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
	slog.Debug("file successfully cropped", "path", result)
	return result, nil
}

func ConvertPdfToPng(fileLocation string) (string, error) {
	dir := filepath.Dir(fileLocation)
	ext := filepath.Ext(fileLocation)
	if ext == ".pdf" {
		result := fmt.Sprintf("%s/converted", dir)
		out, err := exec.Command("pdftoppm", "-singlefile", "-r", "300", "-png", fileLocation, result).CombinedOutput()
		if err != nil {
			return "", errors.New(string(out))
		}
		os.Remove(fileLocation)
		return result + ".png", nil
	}
	return fileLocation, nil
}
