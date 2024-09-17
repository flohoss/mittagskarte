package services

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

const (
	storageLocation  = "storage/"
	downloadLocation = storageLocation + "downloads/"
)

func init() {
	os.MkdirAll(downloadLocation, os.ModePerm)
}

type RestaurantHandler struct {
	restaurants map[string]*Restaurant
	im          *ImageMagic
}

func NewRestaurantHandler(restaurants map[string]*Restaurant) *RestaurantHandler {
	r := &RestaurantHandler{
		restaurants: restaurants,
		im:          NewimageMagic(),
	}
	r.getImageUrls()
	return r
}

func (r *RestaurantHandler) Close() {
	r.im.Close()
}

func (r *RestaurantHandler) getImageUrls() {
	for id := range r.restaurants {
		if r.restaurants[id].PageUrl == "" {
			slog.Debug("no page url, nothing to do...", "id", id)
			continue
		}

		filePath := downloadLocation + id + ".webp"
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			r.restaurants[id].ImageUrl = filePath
			slog.Debug("file already exists, skipping...", "filePath", filePath)
			continue
		}

		cdp := NewChromeDp()
		if err := cdp.NavigateToFinalUrl(r.restaurants[id].PageUrl, &r.restaurants[id].FinalUrl, r.restaurants[id].Parse); err != nil {
			slog.Error("cannot get final url", "err", err)
			cdp.Close()
			continue
		}

		if r.restaurants[id].Parse.IsFile {
			if err := r.handleFile(filePath, id); err != nil {
				slog.Error("cannot handle file", "id", id, "err", err)
				cdp.Close()
				continue
			}
		} else {
			if err := r.handleScreenshot(filePath, id, cdp); err != nil {
				slog.Error("cannot handle screenshot", "id", id, "err", err)
				cdp.Close()
				continue
			}
		}
		cdp.Close()
		r.restaurants[id].ImageUrl = filePath
	}
}

func (r *RestaurantHandler) handleFile(filePath string, id string) error {
	downloadPath, err := fetch.DownloadFile(id, r.restaurants[id].FinalUrl)
	if err != nil {
		return fmt.Errorf("cannot download from '%s': %w", r.restaurants[id].FinalUrl, err)
	}
	if err := r.im.ConvertToWebp(downloadPath, filePath); err != nil {
		return fmt.Errorf("cannot convert to webp: %w", err)
	}
	slog.Info("file downloaded", "filePath", filePath)
	return nil
}

func (r *RestaurantHandler) handleScreenshot(filePath string, id string, cdp *ChromeDp) error {
	if err := cdp.Screenshot(r.restaurants[id].FinalUrl, filePath, r.restaurants[id].Parse.Scan.Chrome, r.restaurants[id].Parse); err != nil {
		return fmt.Errorf("cannot make screenshot: %w", err)
	}
	if err := r.im.Crop(filePath, r.restaurants[id].Parse.Scan.Crop); err != nil {
		return fmt.Errorf("cannot crop screenshot: %w", err)
	}
	slog.Info("screenshot saved", "filePath", filePath)
	return nil
}

func (r *RestaurantHandler) GetAllRestaurants() map[string]*Restaurant {
	return r.restaurants
}

func (r *RestaurantHandler) GetRestaurant(id string) (*Restaurant, error) {
	restaurant, ok := r.restaurants[id]
	if !ok {
		return nil, errors.New("restaurant not found")
	}
	return restaurant, nil
}
