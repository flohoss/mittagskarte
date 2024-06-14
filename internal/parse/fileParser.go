package parse

import (
	"log/slog"
	"os"
	"path/filepath"

	"code.sajari.com/docconv"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch/files"
)

const PublicLocation = "storage/public/menus/"

func init() {
	os.MkdirAll(PublicLocation, os.ModePerm)
}

type FileParser struct {
	id             string
	fileUrl        string
	httpVersion    config.HTTPVersion
	IsNew          bool
	DownloadedFile string
	FileContent    string
}

func NewFileParser(id string, fileUrl string, httpVersion config.HTTPVersion) *FileParser {
	p := &FileParser{
		id:          id,
		fileUrl:     fileUrl,
		httpVersion: httpVersion,
	}
	err := p.download()
	if err != nil {
		return p
	}
	err = p.renameDownloadedFile()
	if err != nil {
		return p
	}
	p.checkIfNew()
	if !p.IsNew {
		return p
	}
	p.parse()
	p.moveToPublicFolder()
	return p
}

func (p *FileParser) download() error {
	var err error
	p.DownloadedFile, err = fetch.DownloadFile(p.id, p.fileUrl, p.httpVersion)
	if err != nil {
		slog.Error("could not download file", "url", p.fileUrl, "err", err)
		return err
	}
	slog.Debug("downloaded file", "file", p.DownloadedFile)
	return nil
}

func (p *FileParser) checkIfNew() {
	newHash, err := files.GenerateHash(p.DownloadedFile)
	if err != nil {
		p.IsNew = true
		return
	}

	_, name, ext := files.GetPathInformation(p.DownloadedFile)
	publicFile := filepath.Join(PublicLocation, name, name+ext)
	oldHash, err := files.GenerateHash(publicFile)
	if err != nil {
		p.IsNew = true
		return
	}

	if oldHash == newHash {
		slog.Debug("file is new", "oldHash", oldHash, "newHash", newHash, "file", p.DownloadedFile)
		p.IsNew = false
		return
	}
	slog.Debug("file is not new", "oldHash", oldHash, "newHash", newHash, "file", p.DownloadedFile)
	p.IsNew = true
}

func (p *FileParser) moveToPublicFolder() error {
	_, name, ext := files.GetPathInformation(p.DownloadedFile)
	publicFile := filepath.Join(PublicLocation, name, name+ext)
	os.MkdirAll(filepath.Join(PublicLocation, name), os.ModePerm)
	err := os.Rename(p.DownloadedFile, publicFile)
	if err != nil {
		slog.Error("could not move file to public folder", "file", p.DownloadedFile, "public", publicFile)
		return err
	}
	slog.Debug("moved file to public folder", "file", p.DownloadedFile, "public", publicFile)
	return nil
}

func (p *FileParser) renameDownloadedFile() error {
	dir, _, ext := files.GetPathInformation(p.DownloadedFile)
	newName := p.id + ext
	newPath := filepath.Join(dir, newName)
	err := os.Rename(p.DownloadedFile, newPath)
	if err != nil {
		slog.Error("could not rename file", "file", p.DownloadedFile)
		return err
	}
	slog.Debug("renamed file", "file", p.DownloadedFile)
	p.DownloadedFile = newPath
	return nil
}

func (p *FileParser) parse() {
	ocr, err := docconv.ConvertPath(p.DownloadedFile)
	if err != nil {
		slog.Error("could not parse file", "file", p.DownloadedFile)
		return
	}
	slog.Debug("parsed file", "file", p.DownloadedFile)
	p.FileContent = ocr.Body
}
