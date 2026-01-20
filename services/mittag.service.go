package services

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/download"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
)

const (
	DownloadsFolder     = "downloads/"
	FinalDownloadFolder = "config/" + DownloadsFolder
	TempDownloadFolder  = "tmp/" + DownloadsFolder
)

func init() {
	os.MkdirAll(FinalDownloadFolder, os.ModePerm)
	os.MkdirAll(TempDownloadFolder, os.ModePerm)
}

type Mittag struct {
	restaurants map[string]*config.Restaurant
	im          *ImageMagic
	cron        *cron.Cron
}

func NewMittag(restaurants map[string]*config.Restaurant) *Mittag {
	r := &Mittag{
		restaurants: restaurants,
		im:          NewimageMagic(),
		cron:        cron.New(),
	}
	for id := range restaurants {
		if restaurants[id].Parse.UpdateCron == "" {
			continue
		}
		id, err := r.cron.AddFunc(restaurants[id].Parse.UpdateCron, func() {
			if err := r.GetImageUrl(restaurants[id], true); err != nil {
				slog.Error(err.Error())
			}
		})
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		slog.Debug("added cron job", "id", id, "schedule", r.cron.Entry(id).Schedule)
	}
	r.cron.Start()
	go r.getImageUrls(false)

	return r
}

func (r *Mittag) Close() {
	if r.im != nil {
		r.im.Close()
	}
}

func (r *Mittag) getImageUrls(overwrite bool) {
	ps, err := newPlaywrightService()
	if err != nil {
		return
	}
	defer ps.close()
	for id := range r.restaurants {
		if err := r.doGetImageUrl(ps, r.restaurants[id], overwrite); err != nil {
			slog.Error(err.Error())
			continue
		}
	}
}

func (r *Mittag) GetImageUrl(restaurant *config.Restaurant, overwrite bool) error {
	restaurant.SetLoading(true)
	defer restaurant.SetLoading(false)
	ps, err := newPlaywrightService()
	if err != nil {
		return err
	}
	defer ps.close()
	if err := r.doGetImageUrl(ps, restaurant, overwrite); err != nil {
		return err
	}
	return nil
}

func (r *Mittag) doGetImageUrl(ps *PlaywrightService, restaurant *config.Restaurant, overwrite bool) error {
	filePath := FinalDownloadFolder + restaurant.ID + ".webp"
	i, err := os.Stat(filePath)
	if !overwrite && !os.IsNotExist(err) {
		slog.Debug("file already exists, skipping...", "filePath", filePath)
		config.SetMenu(filePath, i.ModTime(), restaurant.ID, overwrite)
		return nil
	}

	if r.restaurants[restaurant.ID].PageUrl == "" {
		slog.Debug("no page url, nothing to do...", "id", restaurant.ID)
		return nil
	}

	if r.restaurants[restaurant.ID].Parse.UpdateCron == "" {
		slog.Debug("no parse config, nothing to do...", "id", restaurant.ID)
		return nil
	}

	slog.Debug("getting image url", "id", restaurant.ID)
	tmpPath := ""
	if r.restaurants[restaurant.ID].Parse.DirectDownload != "" {
		tmpPath, err = download.Curl(TempDownloadFolder+restaurant.ID+".pdf", r.restaurants[restaurant.ID].Parse.DirectDownload)
	} else {
		tmpPath, err = ps.doScrape(r.restaurants[restaurant.ID].PageUrl, &r.restaurants[restaurant.ID].Parse)
	}
	if err != nil {
		return err
	}

	err = r.convertToWebp(restaurant.ID, tmpPath, filePath, false)
	if err != nil {
		return err
	}

	os.Remove(tmpPath)

	err = r.im.Trim(filePath)
	if err != nil {
		return err
	}

	i, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		config.SetMenu(filePath, i.ModTime(), restaurant.ID, overwrite)
	}
	return nil
}

func (r *Mittag) convertToWebp(id, tmpPath, filePath string, pdfOverwrite bool) error {
	var err error
	if r.restaurants[id].Parse.FileType == config.PDF || pdfOverwrite {
		err = convertPdfToWebp(tmpPath, filePath)
	} else {
		err = r.im.ConvertToWebp(tmpPath, filePath)
	}
	if err != nil {
		return err
	}
	if err := r.im.ResizeWebp(filePath, filePath); err != nil {
		return err
	}
	return nil
}

func (r *Mittag) UploadMenu(ctx echo.Context, restaurant *config.Restaurant, file *multipart.FileHeader) error {
	restaurant.SetLoading(true)
	defer restaurant.SetLoading(false)
	ext := filepath.Ext(file.Filename)
	allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".webp"}
	if !contains(allowedExtensions, ext) {
		return fmt.Errorf("ungültige Dateierweiterung, erlaubt sind %s", strings.Join(allowedExtensions, ", "))
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	rawPath := filepath.Join(TempDownloadFolder, restaurant.ID) + file.Filename
	dst, err := os.Create(rawPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	filePath := filepath.Join(FinalDownloadFolder, restaurant.ID+".webp")
	if err := r.convertToWebp(restaurant.ID, rawPath, filePath, ext == ".pdf"); err != nil {
		return fmt.Errorf("die Datei kann nicht in das Format .webp konvertiert werden")
	}

	config.SetMenu(filePath, time.Now(), restaurant.ID, true)
	return nil
}

func contains(haistack []string, needle string) bool {
	for _, hai := range haistack {
		if hai == needle {
			return true
		}
	}
	return false
}
