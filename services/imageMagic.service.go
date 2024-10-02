package services

import (
	"fmt"
	"log/slog"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type ImageMagic struct {
}

func NewimageMagic() *ImageMagic {
	imagick.Initialize()
	return &ImageMagic{}
}

func (ic *ImageMagic) Close() {
	if imagick.Terminate != nil {
		imagick.Terminate()
	}
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
	mw.SetCompressionQuality(100)
	if err := mw.WriteImage(newFilePath); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}
	return nil
}
