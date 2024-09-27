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
	storageLocation  = "storage/"
	downloadLocation = storageLocation + "downloads/"
	rawLocation      = downloadLocation + "raw/"
)

func init() {
	os.MkdirAll(rawLocation, os.ModePerm)
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
	go r.getImageUrls()
	return r
}

func (r *Mittag) Close() {
	if r.im != nil {
		r.im.Close()
	}
}

func (r *Mittag) handleRestaurant(s *ScraperService, id string, rawPath string, filePath string) {
	defer s.Close()

	s.navigateToFirstPage(r.restaurants[id].PageUrl)
	if r.restaurants[id].Parse.IsFile {
		if err := s.downloadFile(r.restaurants[id].PageUrl, rawPath, r.restaurants[id].Parse); err != nil {
			s.err <- err
			return
		}
	} else {
		if err := s.screenshot(r.restaurants[id].PageUrl, rawPath, r.restaurants[id].Parse); err != nil {
			s.err <- err
			return
		}
		if err := r.im.Crop(rawPath, r.restaurants[id].Parse.Scan.Crop); err != nil {
			s.err <- err
			return
		}
	}

	if err := r.convertFinalWebp(id, rawPath, filePath); err != nil {
		slog.Error("cannot convert file to final webp", "id", id, "err", err)
		return
	}
	r.restaurants[id].ImageUrl = filePath
	slog.Info("finished", "id", id, "filePath", filePath)
	s.err <- nil
}

func (r *Mittag) getImageUrls() {
	for id := range r.restaurants {
		if r.restaurants[id].PageUrl == "" {
			slog.Debug("no page url, nothing to do...", "id", id)
			continue
		}

		rawPath := rawLocation + id
		filePath := downloadLocation + id + ".webp"
		if _, err := os.Stat(rawPath); !os.IsNotExist(err) {
			slog.Debug("file already exists, skipping...", "filePath", filePath)
			r.restaurants[id].ImageUrl = filePath
			continue
		} else {
			p, err := newPlaywrightService(SiteOptions{
				url:      r.restaurants[id].PageUrl,
				id:       id,
				parse:    &r.restaurants[id].Parse,
				rawPath:  rawPath,
				filePath: filePath,
			})
			if err != nil {
				slog.Error("failed to handle restaurant", "id", id, "err", err)
				p.close()
				continue
			}
			if err := p.doScrape(); err != nil {
				slog.Error("failed to handle restaurant", "id", id, "err", err)
				p.close()
				continue
			}
			p.close()
		}
	}
	slog.Info("all done!")
}

func (r *Mittag) convertFinalWebp(id string, rawPath, filePath string) error {
	var err error
	if r.restaurants[id].Parse.PDF {
		err = convertPdfToWebp(rawPath, filePath)
	} else {
		err = r.im.ConvertToWebp(rawPath, filePath)
	}
	if err != nil {
		return err
	}
	if err := r.im.Trim(filePath); err != nil {
		return err
	}
	return nil
}

func (r *Mittag) GetAllRestaurants() map[string]*Restaurant {
	return r.restaurants
}

func (r *Mittag) GetRestaurant(ctx echo.Context) error {
	restaurant, ok := r.restaurants[ctx.Param("id")]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "Can not find ID")
	}

	return ctx.JSON(http.StatusOK, restaurant)
}

func (r *Mittag) UploadMenu(ctx echo.Context) error {
	restaurant, ok := r.restaurants[ctx.Param("id")]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "Can not find ID")
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided")
	}
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot open file")
	}
	defer src.Close()

	rawPath := filepath.Join(rawLocation, restaurant.ID)
	dst, err := os.Create(rawPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot create file")
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot copy file")
	}
	if err := r.im.ConvertToWebp(rawPath, filepath.Join(downloadLocation, restaurant.ID+".webp")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot convert to webp")
	}
	return ctx.NoContent(http.StatusOK)
}
