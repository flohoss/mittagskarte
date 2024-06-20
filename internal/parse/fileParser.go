package parse

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"code.sajari.com/docconv/client"

	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/pgk/convert"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

const PublicLocation = "storage/menus/"

func init() {
	os.MkdirAll(PublicLocation, os.ModePerm)
}

type FileParser struct {
	OutputFileLocation string
	OutputFileContent  string
}

type OCRResponse struct {
	Result string `json:"result"`
}

func NewFileParser(id string, fileUrl string, httpVersion config.HTTPVersion, needsParsing bool) *FileParser {
	downloadedFile, err := fetch.DownloadFile(id, fileUrl, httpVersion)
	if err != nil {
		slog.Error("could not download file", "url", fileUrl, "err", err)
		return nil
	}

	base := filepath.Base(downloadedFile)
	publicFile := filepath.Join(PublicLocation, base)
	os.Rename(downloadedFile, publicFile)

	ocr := ""
	if needsParsing {
		ocr, err = requestOCR(publicFile)
		if err != nil {
			slog.Error("could not parse file", "file", publicFile, "err", err)
			return nil
		}
	}

	outputFileLocation, err := convert.ConvertToWebP(publicFile, false)
	if err != nil {
		slog.Error("could not convert file to webp", "file", publicFile, "err", err)
		return nil
	}

	return &FileParser{
		OutputFileLocation: outputFileLocation,
		OutputFileContent:  ocr,
	}
}

func requestOCR(fileLocation string) (string, error) {
	ext := filepath.Ext(fileLocation)
	switch ext {
	case ".pdf":
		c := client.New(client.WithEndpoint("docd:8888"))
		res, err := client.ConvertPath(c, fileLocation)
		if err != nil {
			return "", err
		}
		return res.Body, nil
	case ".jpg", ".jpeg", ".png":
		return makeOcrResuest(fileLocation)
	}
	return "", nil
}

func makeOcrResuest(fileLocation string) (string, error) {
	file, err := os.Open(fileLocation)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", fileLocation)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://ocrserver:8080/file", &requestBody)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var ocrResponse OCRResponse
	json.Unmarshal(body, &ocrResponse)
	return ocrResponse.Result, nil
}
