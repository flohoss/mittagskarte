package services

import (
	"fmt"
	"log/slog"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type ImageMagic struct {
	compressionQuality uint
	maximumWidth       uint
}

func NewimageMagic() *ImageMagic {
	imagick.Initialize()
	return &ImageMagic{
		compressionQuality: 90,
		maximumWidth:       1920,
	}
}

func (ic *ImageMagic) Close() {
	if imagick.Terminate != nil {
		imagick.Terminate()
	}
}

func (ic *ImageMagic) ResizeWebp(srcPath, dstPath string) error {
	slog.Debug("resizing webp", "path", srcPath)
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	if err := mw.ReadImage(srcPath); err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

	if width <= ic.maximumWidth {
		slog.Debug("image below max width, skipping resize and save", "width", width, "maxWidth", ic.maximumWidth)
		return nil
	}

	newHeight := uint(float64(height) * float64(ic.maximumWidth) / float64(width))
	if err := mw.ResizeImage(ic.maximumWidth, newHeight, imagick.FILTER_LANCZOS, 1); err != nil {
		return fmt.Errorf("failed to resize image: %w", err)
	}

	mw.SetCompressionQuality(ic.compressionQuality)
	if err := mw.WriteImage(dstPath); err != nil {
		return fmt.Errorf("failed to write resized image: %w", err)
	}
	return nil
}

func (ic *ImageMagic) ConvertToWebp(oldFilePath string, newFilePath string) error {
	slog.Debug("converting image to webp", "path", oldFilePath)
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	if err := mw.ReadImage(oldFilePath); err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}
	if err := mw.SetImageFormat("webp"); err != nil {
		return fmt.Errorf("failed to set image format webp: %w", err)
	}

	mw.SetCompressionQuality(ic.compressionQuality)
	if err := mw.WriteImage(newFilePath); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}
	return nil
}

func (ic *ImageMagic) Trim(filePath string) error {
	slog.Debug("trimming image", "path", filePath)
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	if err := mw.ReadImage(filePath); err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}
	mw.SetImageBackgroundColor(imagick.NewPixelWand())
	if err := mw.TrimImage(0); err != nil {
		return fmt.Errorf("failed to trim image: %w", err)
	}
	if err := mw.WriteImage(filePath); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}
	return nil
}
