package parse

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"code.sajari.com/docconv"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/pgk/convert"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

const PublicLocation = "public/menus/"

func init() {
	os.MkdirAll(PublicLocation, os.ModePerm)
}

type FileParser struct {
	OutputFileLocation string
	OutputFileContent  string
}

func NewFileParser(id string, fileUrl string, httpVersion config.HTTPVersion) *FileParser {
	downloadedFile, err := fetch.DownloadFile(id, fileUrl, httpVersion)
	if err != nil {
		slog.Error("could not download file", "url", fileUrl, "err", err)
		return nil
	}
	fmt.Println(downloadedFile)

	base := filepath.Base(downloadedFile)
	publicFile := filepath.Join(PublicLocation, base)
	os.Rename(downloadedFile, publicFile)

	outputFileLocation, err := convert.ConvertToWebP(publicFile, false)
	if err != nil {
		slog.Error("could not convert file to webp", "file", publicFile, "err", err)
		return nil
	}

	ocr, err := docconv.ConvertPath(outputFileLocation)
	if err != nil {
		slog.Error("could not parse file", "file", outputFileLocation, "err", err)
		return nil
	}

	return &FileParser{
		OutputFileLocation: outputFileLocation,
		OutputFileContent:  ocr.Body,
	}
}
