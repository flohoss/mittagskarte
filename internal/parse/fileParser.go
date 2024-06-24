package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"code.sajari.com/docconv/client"

	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
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

func NewFileParser(id string, fileUrl string, httpVersion config.HTTPVersion, needsParsing bool, env *env.Env) *FileParser {
	downloadedFile, err := fetch.DownloadFile(id, fileUrl, httpVersion)
	if err != nil {
		slog.Error("could not download file", "url", fileUrl, "err", err)
		return nil
	}

	ocr, outputFileLocation := MoveAndParse(downloadedFile, needsParsing, env)

	return &FileParser{
		OutputFileLocation: outputFileLocation,
		OutputFileContent:  ocr,
	}
}

func MoveAndParse(downloadedFile string, needsParsing bool, env *env.Env) (string, string) {
	base := filepath.Base(downloadedFile)
	publicFile := filepath.Join(PublicLocation, base)
	os.Rename(downloadedFile, publicFile)

	var err error
	ocr := ""
	if needsParsing {
		ocr, err = requestOCR(publicFile, env)
		if err != nil {
			slog.Error("could not parse file", "file", publicFile, "err", err)
		}
	}

	outputFileLocation, err := convert.ConvertToWebP(publicFile, false)
	if err != nil {
		slog.Error("could not convert file to webp", "file", publicFile, "err", err)
	}
	return ocr, outputFileLocation
}

func requestOCR(fileLocation string, env *env.Env) (string, error) {
	ext := filepath.Ext(fileLocation)
	switch ext {
	case ".pdf":
		c := client.New(client.WithEndpoint(fmt.Sprintf("%s:%d", env.DocHost, env.DocPort)))
		res, err := client.ConvertPath(c, fileLocation)
		if err != nil {
			return "", err
		}
		return res.Body, nil
	case ".jpg", ".jpeg", ".png":
		return makeOcrRequest(fileLocation, env.OCRHost, env.OCRPort)
	}
	return "", nil
}

func makeOcrRequest(fileLocation string, host string, port int) (string, error) {
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

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/file", host, port), &requestBody)
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
