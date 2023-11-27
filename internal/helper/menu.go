package helper

import (
	"code.sajari.com/docconv"
)

func ParseMenu(filePath string) (string, error) {
	ocr, err := docconv.ConvertPath(filePath)
	if err != nil {
		return "", err
	}
	return ocr.Body, nil
}
