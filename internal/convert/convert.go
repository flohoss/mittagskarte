package convert

import (
	"errors"
	"fmt"
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

func CreateWebp(fileLocation string) (string, error) {
	newFile := ReplaceEndingToWebp(fileLocation)
	app := "cwebp"
	args := []string{fileLocation, "-o", newFile}
	cmd := exec.Command(app, args...)
	err := cmd.Run()
	if err != nil {
		return newFile, errors.New("cannot convert to webp")
	}
	os.Remove(fileLocation)
	return newFile, nil
}

func ConvertPdfToPng(fileLocation string, resultName string, dpi string) (string, error) {
	dir := filepath.Dir(fileLocation)
	result := fmt.Sprintf("%s/%s", dir, resultName)
	if filepath.Ext(fileLocation) != ".pdf" {
		os.Rename(fileLocation, result+".png")
	} else {
		app := "pdftoppm"
		args := []string{"-png", "-singlefile", "-r", dpi, fileLocation, result}
		cmd := exec.Command(app, args...)
		err := cmd.Run()
		if err != nil {
			return "", err
		}
		os.Remove(fileLocation)
	}
	return result + ".png", nil
}
