package services

import (
	"errors"
	"log/slog"
	"os"
)

const (
	storageLocation  = "storage/"
	downloadLocation = storageLocation + "downloads/"
	rawLocation      = downloadLocation + "raw/"
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

		rawPath := rawLocation + id
		filePath := downloadLocation + id + ".webp"
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			r.restaurants[id].ImageUrl = filePath
			slog.Debug("file already exists, skipping...", "filePath", filePath)
			continue
		}

		cdp := NewScraper()
		if r.restaurants[id].Parse.IsFile {
			if err := cdp.DownloadFile(r.restaurants[id].PageUrl, rawPath, r.restaurants[id].Parse); err != nil {
				slog.Error("cannot handle file", "id", id, "err", err)
				cdp.Close()
				continue
			}
			if err := r.im.ConvertToWebp(rawPath, filePath); err != nil {
				slog.Error("cannot convert to webp", "id", id, "err", err)
				cdp.Close()
				continue
			}
		} else {
			if err := cdp.Screenshot(r.restaurants[id].PageUrl, rawPath, r.restaurants[id].Parse); err != nil {
				slog.Error("cannot take screenshot", "id", id, "err", err)
				cdp.Close()
				continue
			}
			if err := r.im.ConvertToWebp(rawPath, filePath); err != nil {
				slog.Error("cannot convert to webp", "id", id, "err", err)
				cdp.Close()
				continue
			}
			if err := r.im.Crop(filePath, r.restaurants[id].Parse.Scan.Crop); err != nil {
				slog.Error("cannot crop image", "id", id, "err", err)
				cdp.Close()
				continue
			}
		}
		cdp.Close()
		r.restaurants[id].ImageUrl = filePath
	}
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
