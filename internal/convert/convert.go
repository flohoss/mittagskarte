package convert

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

func ConvertPdfToPng(fileLocation string, dpi string) (string, error) {
	app := "pdftoppm"
	args := []string{"-png", "-singlefile", "-aa", "yes", "-r", dpi, fileLocation, fileLocation}
	cmd := exec.Command(app, args...)
	err := cmd.Run()
	if err != nil {
		zap.L().Error("Cannot convert PDF to PNG", zap.Error(err))
		return "", err
	}
	os.Remove(fileLocation)
	return fileLocation + ".png", nil
}

func CropPng(fileLocation string, crop string, trim bool) {
	app := "convert"
	args := []string{fileLocation, "-crop", crop}
	if trim {
		args = append(args, "-trim", fileLocation)
	} else {
		args = append(args, fileLocation)
	}
	cmd := exec.Command(app, args...)
	err := cmd.Run()
	if err != nil {
		zap.L().Error("Cannot crop PNG", zap.Error(err))
	}
}

func ReplaceEndingToWebp(fileLocation string) string {
	newFile := strings.Replace(fileLocation, ".png", ".webp", 1)
	if strings.Contains(fileLocation, "jpg") {
		newFile = strings.Replace(fileLocation, ".jpg", ".webp", 1)
	}
	return newFile
}

func CreateWebp(fileLocation string) string {
	newFile := ReplaceEndingToWebp(fileLocation)
	app := "convert"
	args := []string{fileLocation, newFile}
	cmd := exec.Command(app, args...)
	err := cmd.Run()
	if err != nil {
		zap.L().Error("Cannot create WEBP", zap.Error(err))
	}
	return newFile
}

/*
 * https://tesseract-ocr.github.io/tessdoc/Command-Line-Usage.html
 *
 * --psm 3
 *	-> Fully automatic page segmentation, but no OSD
 * --psm 6
 * 	-> Assume a single uniform block of text
 * --psm 6 -c preserve_interword_spaces=1
 *	-> Use -c preserve_interword_spaces=1 to preserve spaces
 * –-psm 11
 *	-> Use pdftotext for preserving layout for text output
 */
func OCR(fileLocation string, psm uint) ([]byte, error) {
	app := "tesseract"
	args := []string{fileLocation, "stdout", "-l", "eng+deu", "--psm", fmt.Sprintf("%d", psm)}
	out, err := exec.Command(app, args...).Output()

	if err != nil {
		zap.L().Error("Cannot perform OCR", zap.Error(err))
		return nil, err
	}
	return out, nil
}
