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
	"github.com/flohoss/mittagskarte/internal/scheduler"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
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
	scheduler   *scheduler.Scheduler
	ps          *PlaywrightService
}

func NewMittag(restaurants map[string]*config.Restaurant) *Mittag {
	r := &Mittag{
		restaurants: restaurants,
		im:          NewimageMagic(),
		scheduler:   scheduler.New(),
	}

	var cronJobs = config.GetAllCrons()
	for sTime, restaurants := range cronJobs {
		r.scheduler.Add(sTime, func() {
			r.getImageUrls(restaurants, true)
		})
		var ids []string
		for id := range restaurants {
			ids = append(ids, id)
		}
		slog.Debug("added cron job", "schedule", sTime, "restaurants", strings.Join(ids, ","))
	}

	go r.getImageUrls(nil, false)

	return r
}

func (r *Mittag) getPlaywrightService() (*PlaywrightService, error) {
	if r.ps != nil {
		return r.ps, nil
	}

	ps, err := newPlaywrightService()
	if err != nil {
		slog.Error("could not initialize PlaywrightService", "error", err)
	} else {
		r.ps = ps
	}
	return r.ps, err
}

func (r *Mittag) Close() {
	if r.im != nil {
		r.im.Close()
	}
	if r.ps != nil {
		r.ps.close()
	}
}

func (r *Mittag) getImageUrls(restaurants map[string]*config.Restaurant, overwrite bool) {
	if restaurants == nil {
		restaurants = r.restaurants
	}

	for id := range restaurants {
		if err := r.doGetImageUrl(restaurants[id], overwrite); err != nil {
			slog.Error(err.Error())
			sentry.CaptureException(err)
			continue
		}
	}
}

func (r *Mittag) GetImageUrl(restaurant *config.Restaurant, overwrite bool) error {
	if err := r.doGetImageUrl(restaurant, overwrite); err != nil {
		sentry.CaptureException(err)
		return err
	}
	return nil
}

func (r *Mittag) doGetImageUrl(restaurant *config.Restaurant, overwrite bool) error {
	filePath := FinalDownloadFolder + restaurant.ID + ".webp"
	i, err := os.Stat(filePath)
	if !overwrite && !os.IsNotExist(err) {
		slog.Debug("file already exists, skipping...", "filePath", filePath)
		config.SetMenu(filePath, i.ModTime(), restaurant.ID, overwrite)
		return nil
	}

	restaurant.SetLoading(true)
	defer restaurant.SetLoading(false)

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
	if r.restaurants[restaurant.ID].Parse.DownloadURL != "" {
		tmpPath, err = download.Curl(TempDownloadFolder+restaurant.ID, r.restaurants[restaurant.ID].Parse.DownloadURL)
		if err != nil {
			return err
		}
	} else {
		ps, err := r.getPlaywrightService()
		if err != nil {
			return fmt.Errorf("could not get PlaywrightService: %w", err)
		}
		tmpPath, err = ps.doScrape(r.restaurants[restaurant.ID].PageUrl, &r.restaurants[restaurant.ID].Parse)
		if err != nil {
			return err
		}
	}
	if tmpPath == "" {
		slog.Error("doScrape/Curl returned empty path", "id", restaurant.ID)
		return fmt.Errorf("scraping returned empty file path")
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
	if !contains(config.GetAllowedExtensions(), ext) {
		return fmt.Errorf(config.GetAllowedExtensionsMessage())
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
