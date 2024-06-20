package convert

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

func ConvertToWebP(fileLocation string, keepOriginal bool) (string, error) {
	ext := filepath.Ext(fileLocation)
	outputFileLocation := strings.Replace(fileLocation, ext, ".webp", 1)
	switch ext {
	case ".pdf":
		if err := convertPdfToWebp(fileLocation, outputFileLocation); err != nil {
			return fileLocation, err
		}
	case ".jpg", ".jpeg", ".png":
		if err := convertImageToWebp(fileLocation, outputFileLocation, ext); err != nil {
			return fileLocation, err
		}
	default:
		return fileLocation, fmt.Errorf("unknown file type: %s", ext)
	}

	if !keepOriginal {
		if err := os.Remove(fileLocation); err != nil {
			return fileLocation, err
		}
	}
	return outputFileLocation, nil
}

func convertImageToWebp(fileLocation string, outputFileLocation string, ext string) error {
	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	var img image.Image
	switch ext {
	case ".png":
		img, err = png.Decode(file)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	}
	if err != nil {
		return err
	}
	return encodeWebp(outputFileLocation, img)
}

func convertPdfToWebp(fileLocation string, outputFileLocation string) error {
	doc, err := fitz.New(fileLocation)
	if err != nil {
		return err
	}
	img, err := doc.Image(0)
	if err != nil {
		return err
	}
	return encodeWebp(outputFileLocation, img)
}

func encodeWebp(outputFileLocation string, img image.Image) error {
	output, err := os.Create(outputFileLocation)
	if err != nil {
		return err
	}
	defer output.Close()
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 80)

	if err != nil {
		return err
	}
	if err := webp.Encode(output, img, options); err != nil {
		return err
	}
	return nil
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
