package pdf

import (
	"image"
	"os"

	"github.com/chai2010/webp"
	"github.com/gen2brain/go-fitz"
)

func ConvertToWebp(oldFilePath string, newFilePath string) error {
	doc, err := fitz.New(oldFilePath)
	if err != nil {
		return err
	}
	img, err := doc.Image(0)
	if err != nil {
		return err
	}
	return encodeWebp(newFilePath, img)
}

func encodeWebp(filePath string, img image.Image) error {
	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer output.Close()

	if err := webp.Encode(output, img, &webp.Options{Lossless: true, Quality: 100}); err != nil {
		return err
	}
	return nil
}
