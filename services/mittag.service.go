package services

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
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
}

func NewMittag(restaurants map[string]*Restaurant) *Mittag {
	r := &Mittag{
		restaurants: restaurants,
		im:          NewimageMagic(),
	}
	r.getImageUrls()
	return r
}

func (r *Mittag) Close() {
	if r.im != nil {
		r.im.Close()
	}
}

func (r *Mittag) getImageUrls() {
	p, err := newPlaywrightService()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer p.close()
	for id := range r.restaurants {
		slog.Debug("getting image url", "id", id)

		filePath := FinalDownloadFolder + id + ".webp"
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			slog.Debug("file already exists, skipping...", "filePath", filePath)
			r.restaurants[id].ImageUrl = filePath
			continue
		}

		if r.restaurants[id].PageUrl == "" {
			slog.Debug("no page url, nothing to do...", "id", id)
			continue
		}

		tmpPath, err := p.doScrape(r.restaurants[id].PageUrl, &r.restaurants[id].Parse)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		err = r.convertToWebp(id, tmpPath, filePath, false)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		os.Remove(tmpPath)

		if r.restaurants[id].Parse.FileType != PDF && r.restaurants[id].Parse.FileType != Image {
			err = r.im.Trim(filePath)
			if err != nil {
				slog.Error(err.Error())
				continue
			}
		}
	}
	slog.Info("all done!")
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
		return echo.NewHTTPError(http.StatusNotFound, "Can not find ID")
	}

	return ctx.JSON(http.StatusOK, restaurant.GetCleanRestaurant())
}

func (r *Mittag) UploadMenu(ctx echo.Context) error {
	restaurant, ok := r.restaurants[ctx.Param("id")]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "can not find ID")
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "no file provided")
	}
	ext := filepath.Ext(file.Filename)
	if !contains([]string{".pdf", ".jpg", ".jpeg", ".png", ".webp"}, ext) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid file extension")
	}
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot open file")
	}
	defer src.Close()

	rawPath := filepath.Join(TempDownloadFolder, restaurant.ID)
	dst, err := os.Create(rawPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot create file")
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot copy file")
	}
	filePath := filepath.Join(FinalDownloadFolder, restaurant.ID+".webp")
	if err := r.convertToWebp(restaurant.ID, rawPath, filePath, ext == ".pdf"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot convert to webp")
	}
	restaurant.ImageUrl = filePath
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
