package services

import (
	"os"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type ImageMagic struct {
	mw *imagick.MagickWand
}

func NewimageMagic() *ImageMagic {
	imagick.Initialize()
	mw := imagick.NewMagickWand()
	return &ImageMagic{
		mw: mw,
	}
}

func (ic *ImageMagic) Close() {
	if imagick.Terminate != nil {
		imagick.Terminate()
	}
}

func (ic *ImageMagic) Crop(filePath string, crop Crop) error {
	if err := ic.mw.ReadImage(filePath); err != nil {
		return err
	}
	if err := ic.mw.CropImage(crop.Width, crop.Height, crop.OffsetX, crop.OffsetY); err != nil {
		return err
	}
	if err := ic.mw.WriteImage(filePath); err != nil {
		return err
	}
	return nil
}

func (ic *ImageMagic) ConvertToWebp(oldFilePath string, newFilePath string) error {
	if err := ic.mw.ReadImage(oldFilePath); err != nil {
		return err
	}
	if err := ic.mw.SetImageFormat("webp"); err != nil {
		return err
	}
	if err := ic.mw.WriteImage(newFilePath); err != nil {
		return err
	}
	os.Remove(oldFilePath)
	return nil
}
