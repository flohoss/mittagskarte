package pdf

import (
	"image"
	"os"

	"github.com/chai2010/webp"
	"github.com/gen2brain/go-fitz"
)

func MergeAllPDFPagesToWebp(inputPDFPath, outputWebpPath string) error {
	doc, err := fitz.New(inputPDFPath)
	if err != nil {
		return err
	}
	defer doc.Close()
	pageCount := doc.NumPage()
	if pageCount == 0 {
		return nil
	}

	if pageCount == 1 {
		return FirstPDFPageToWebp(inputPDFPath, outputWebpPath)
	}

	images := make([]image.Image, 0, pageCount)
	for i := 0; i < pageCount; i++ {
		img, err := doc.Image(i)
		if err != nil {
			return err
		}
		images = append(images, img)
	}
	return encodeImageAsWebp(outputWebpPath, mergeImagesVertically(images))
}

func FirstPDFPageToWebp(inputPDFPath, outputWebpPath string) error {
	doc, err := fitz.New(inputPDFPath)
	if err != nil {
		return err
	}
	defer doc.Close()
	img, err := doc.Image(0)
	if err != nil {
		return err
	}
	return encodeImageAsWebp(outputWebpPath, img)
}

func mergeImagesVertically(images []image.Image) image.Image {
	maxWidth, totalHeight := getMaxWidthAndTotalHeight(images)
	out := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))
	currY := 0
	for _, img := range images {
		offsetX := getHorizontalCenterOffset(img, maxWidth)
		copyImageToPosition(out, img, offsetX, currY)
		currY += img.Bounds().Dy()
	}
	return out
}

func getMaxWidthAndTotalHeight(images []image.Image) (int, int) {
	maxWidth := 0
	totalHeight := 0
	for _, img := range images {
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		if w > maxWidth {
			maxWidth = w
		}
		totalHeight += h
	}
	return maxWidth, totalHeight
}

func getHorizontalCenterOffset(img image.Image, maxWidth int) int {
	return (maxWidth - img.Bounds().Dx()) / 2
}

func copyImageToPosition(dst *image.RGBA, src image.Image, offsetX, offsetY int) {
	r := src.Bounds()
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			dst.Set(x+offsetX, y+offsetY, src.At(r.Min.X+x, r.Min.Y+y))
		}
	}
}

func encodeImageAsWebp(filePath string, img image.Image) error {
	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer output.Close()
	return webp.Encode(output, img, &webp.Options{Lossless: true, Quality: 100})
}
