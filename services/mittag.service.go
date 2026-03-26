package services

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/download"
	"github.com/flohoss/mittagskarte/internal/events"
	"github.com/flohoss/mittagskarte/internal/scheduler"
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
	mu        sync.RWMutex
	im        *ImageMagic
	scheduler *scheduler.Scheduler
	ps        *PlaywrightService
	Events    *events.Event
}

func NewMittag() *Mittag {
	r := &Mittag{
		im:        NewimageMagic(),
		scheduler: scheduler.New(),
	}

	cronJobs := config.GetAllCrons()
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

func (r *Mittag) SetEvents(e *events.Event) {
	r.Events = e
}

func (r *Mittag) getImageUrls(restaurants map[string]*config.Restaurant, overwrite bool) {
	if restaurants == nil {
		restaurants = config.GetRestaurants()
	}

	for id := range restaurants {
		if err := r.doGetImageUrl(restaurants[id], overwrite); err != nil {
			slog.Error(err.Error())
			continue
		}
	}
}

func (r *Mittag) GetImageUrl(restaurant *config.Restaurant, overwrite bool) error {
	if err := r.doGetImageUrl(restaurant, overwrite); err != nil {
		return err
	}
	return nil
}

func (r *Mittag) doGetImageUrl(restaurant *config.Restaurant, overwrite bool) error {
	filePath := FinalDownloadFolder + restaurant.ID + ".webp"
	i, err := os.Stat(filePath)
	if !overwrite && !os.IsNotExist(err) {
		slog.Debug("file already exists, skipping...", "filePath", filePath)
		restaurant.SetMenu(filePath, i.ModTime())
		return nil
	}

	restaurant.SetLoading(true)
	defer restaurant.SetLoading(false)

	if restaurant.PageUrl == "" {
		slog.Debug("no page url, nothing to do...", "id", restaurant.ID)
		return nil
	}

	if restaurant.Parse.UpdateCron == "" {
		slog.Debug("no parse config, nothing to do...", "id", restaurant.ID)
		return nil
	}

	slog.Debug("getting image url", "id", restaurant.ID)
	tmpPath := ""
	if restaurant.Parse.DownloadURL != "" {
		tmpPath, err = download.Curl(TempDownloadFolder+restaurant.ID, restaurant.Parse.DownloadURL)
		if err != nil {
			return err
		}
	} else {
		ps, err := r.getPlaywrightService()
		if err != nil {
			return fmt.Errorf("could not get PlaywrightService: %w", err)
		}
		tmpPath, err = ps.doScrape(restaurant.PageUrl, &restaurant.Parse)
		if err != nil {
			return err
		}
	}
	if tmpPath == "" {
		slog.Error("doScrape/Curl returned empty path", "id", restaurant.ID)
		return fmt.Errorf("scraping returned empty file path")
	}

	err = r.convertToWebp(restaurant, tmpPath, filePath, false)
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
		restaurant.SetMenu(filePath, i.ModTime())
	}
	return nil
}

func (r *Mittag) convertToWebp(restaurant *config.Restaurant, tmpPath, filePath string, pdfOverwrite bool) error {
	var convertErr error
	if restaurant.Parse.FileType == config.PDF || pdfOverwrite {
		convertErr = convertPdfToWebp(tmpPath, filePath)
	} else {
		convertErr = r.im.ConvertToWebp(tmpPath, filePath)
	}
	if convertErr != nil {
		return convertErr
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
	if err := r.convertToWebp(restaurant, rawPath, filePath, ext == ".pdf"); err != nil {
		return fmt.Errorf("die Datei kann nicht in das Format .webp konvertiert werden")
	}

	restaurant.SetMenu(filePath, time.Now())
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
