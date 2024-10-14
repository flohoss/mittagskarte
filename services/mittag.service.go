package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
)

const (
	FinalDownloadFolder = "storage/downloads/"
	TempDownloadFolder  = "tmp/downloads/"
)

func init() {
	os.MkdirAll(FinalDownloadFolder, os.ModePerm)
	os.MkdirAll(TempDownloadFolder, os.ModePerm)
}

type Mittag struct {
	restaurants map[string]*Restaurant
	im          *ImageMagic
	cron        *cron.Cron
}

func NewMittag(restaurants map[string]*Restaurant) *Mittag {
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
			if err := r.getImageUrl(restaurants[id], true); err != nil {
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
	r.getImageUrls(false)

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

func (r *Mittag) getImageUrl(restaurant *Restaurant, overwrite bool) error {
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

func (r *Mittag) doGetImageUrl(ps *PlaywrightService, restaurant *Restaurant, overwrite bool) error {
	slog.Debug("getting image url", "id", restaurant.ID)

	filePath := FinalDownloadFolder + restaurant.ID + ".webp"
	_, err := os.Stat(filePath)
	if !overwrite && !os.IsNotExist(err) {
		slog.Debug("file already exists, skipping...", "filePath", filePath)
		r.restaurants[restaurant.ID].ImageUrl = filePath
		return nil
	}

	if r.restaurants[restaurant.ID].PageUrl == "" {
		slog.Debug("no page url, nothing to do...", "id", restaurant.ID)
		return nil
	}

	tmpPath, err := ps.doScrape(r.restaurants[restaurant.ID].PageUrl, &r.restaurants[restaurant.ID].Parse)
	if err != nil {
		return err
	}

	err = r.convertToWebp(restaurant.ID, tmpPath, filePath, false)
	if err != nil {
		return err
	}

	os.Remove(tmpPath)

	if r.restaurants[restaurant.ID].Parse.FileType != PDF && r.restaurants[restaurant.ID].Parse.FileType != Image {
		err = r.im.Trim(filePath)
		if err != nil {
			return err
		}
	}

	r.restaurants[restaurant.ID].ImageUrl = filePath
	return nil
}

func (r *Mittag) convertToWebp(id, tmpPath, filePath string, pdfOverwrite bool) error {
	var err error
	if r.restaurants[id].Parse.FileType == PDF || pdfOverwrite {
		err = convertPdfToWebp(tmpPath, filePath)
	} else {
		err = r.im.ConvertToWebp(tmpPath, filePath)
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Mittag) GetAllRestaurants(ctx echo.Context) error {
	apiResponse := make(map[string]*CleanRestaurant)
	for key, restaurant := range r.restaurants {
		apiResponse[key] = restaurant.GetCleanRestaurant()
	}
	return ctx.JSON(http.StatusOK, apiResponse)
}

func (r *Mittag) GetRestaurant(ctx echo.Context) error {
	restaurant, ok := r.restaurants[ctx.Param("id")]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "ID konnte nicht gefunden werden")
	}

	return ctx.JSON(http.StatusOK, restaurant.GetCleanRestaurant())
}

func (r *Mittag) UploadMenu(ctx echo.Context) error {
	restaurant, ok := r.restaurants[ctx.Param("id")]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "ID konnte nicht gefunden werden")
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "keine Datei vorhanden")
	}
	ext := filepath.Ext(file.Filename)
	allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".webp"}
	if !contains(allowedExtensions, ext) {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ungültige Dateierweiterung, erlaubt sind %s", strings.Join(allowedExtensions, ", ")))
	}
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "die Datei kann nicht geöffnet werden")
	}
	defer src.Close()

	rawPath := filepath.Join(TempDownloadFolder, restaurant.ID)
	dst, err := os.Create(rawPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "die Datei kann auf dem Server nicht erstellt werden")
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "die Datei kann nicht kopiert werden")
	}
	filePath := filepath.Join(FinalDownloadFolder, restaurant.ID+".webp")
	if err := r.convertToWebp(restaurant.ID, rawPath, filePath, ext == ".pdf"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "die Datei kann nicht in das Format .webp konvertiert werden")
	}
	restaurant.ImageUrl = filePath
	return ctx.JSON(http.StatusOK, restaurant.GetCleanRestaurant())
}

func (r *Mittag) UpdateRestaurant(ctx echo.Context) error {
	restaurant, ok := r.restaurants[ctx.Param("id")]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "ID konnte nicht gefunden werden")
	}

	if restaurant.PageUrl == "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("für %s ist keine Speisekarte online, bitte laden Sie manuell eine Speisekarte von diesem Restaurant hoch", restaurant.ID))
	}

	if err := r.getImageUrl(restaurant, true); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, restaurant.GetCleanRestaurant())
}

func contains(haistack []string, needle string) bool {
	for _, hai := range haistack {
		if hai == needle {
			return true
		}
	}
	return false
}
