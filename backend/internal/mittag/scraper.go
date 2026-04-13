package mittag

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flohoss/mittagskarte/internal/image"
	"github.com/flohoss/mittagskarte/internal/pdf"
	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/flohoss/mittagskarte/internal/web"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
)

type restaurantsProvider func(app core.App) ([]*restaurant.Restaurant, error)

const (
	ScrapeStatusUpdating = "updating"
	ScrapeStatusQueued   = "queued"
	ScrapeStatusCooldown = "cooldown"
	ScrapeStatusIdle     = "idle"

	RestaurantsStatusTopic = "restaurants/status"
)

type Scraper struct {
	app            core.App
	web            *web.Web
	im             *image.ImageMagic
	getRestaurants restaurantsProvider

	queueMu     sync.Mutex
	queueCond   *sync.Cond
	scrapeQueue []string
	queued      map[string]struct{}
	inFlight    map[string]struct{}
	lastRunAt   map[string]time.Time
	cooldown    time.Duration
	queueClosed bool
	workerWg    sync.WaitGroup
}

func NewScraper(app core.App, webService *web.Web, imageMagic *image.ImageMagic, provider restaurantsProvider) *Scraper {
	s := &Scraper{
		app:            app,
		web:            webService,
		im:             imageMagic,
		getRestaurants: provider,
		queued:         make(map[string]struct{}),
		inFlight:       make(map[string]struct{}),
		lastRunAt:      make(map[string]time.Time),
		cooldown:       5 * time.Minute,
	}

	s.queueCond = sync.NewCond(&s.queueMu)
	s.workerWg.Add(1)
	go s.runScrapeWorker()

	return s
}

func (s *Scraper) Close() {
	s.queueMu.Lock()
	s.queueClosed = true
	s.queueMu.Unlock()
	s.queueCond.Broadcast()
	s.workerWg.Wait()

	s.web.Close()
	s.im.Close()
}

func (s *Scraper) Enqueue(restaurants []*restaurant.Restaurant) {
	var err error

	if restaurants == nil {
		restaurants, err = s.getRestaurants(s.app)
		if err != nil {
			s.app.Logger().Error("Error fetching restaurants", "error", err)
			return
		}
	}

	s.queueMu.Lock()

	if s.queueClosed {
		s.queueMu.Unlock()
		s.app.Logger().Warn("Scrape queue is closed, skipping enqueue")
		return
	}

	now := time.Now()
	queuedRestaurantIDs := make([]string, 0)
	for _, r := range restaurants {
		if _, ok := s.queued[r.ID]; ok {
			s.app.Logger().Debug("Skipping enqueue: restaurant already queued", "id", r.ID)
			continue
		}
		if _, ok := s.inFlight[r.ID]; ok {
			s.app.Logger().Debug("Skipping enqueue: restaurant scrape in progress", "id", r.ID)
			continue
		}
		if lastRunAt, ok := s.lastRunAt[r.ID]; ok && now.Sub(lastRunAt) < s.cooldown {
			s.app.Logger().Debug("Skipping enqueue: restaurant in cooldown", "id", r.ID, "nextAllowedAt", lastRunAt.Add(s.cooldown))
			continue
		}

		s.scrapeQueue = append(s.scrapeQueue, r.ID)
		s.queued[r.ID] = struct{}{}
		queuedRestaurantIDs = append(queuedRestaurantIDs, r.ID)
	}

	s.queueCond.Broadcast()
	s.queueMu.Unlock()

	for _, restaurantID := range queuedRestaurantIDs {
		s.notifyRestaurantStatus(restaurantID)
	}
}

func (s *Scraper) runScrapeWorker() {
	defer s.workerWg.Done()

	for {
		restaurantID, ok := s.dequeueScrapeID()
		if !ok {
			return
		}

		s.notifyRestaurantStatus(restaurantID)

		r, err := s.getRestaurantByID(restaurantID)
		if err != nil {
			s.app.Logger().Error("Error resolving restaurant for scrape", "id", restaurantID, "error", err)
			s.markScrapeDone(restaurantID)
			s.notifyRestaurantStatus(restaurantID)
			continue
		}

		scrapeErr := s.scrapeSingle(r)
		if errors.Is(scrapeErr, restaurant.ErrManualUploadOnly) {
			s.app.Logger().Warn("Skipped automated update: manual upload only", "name", r.Name)
		} else if errors.Is(scrapeErr, restaurant.ErrMenuUnchanged) {
			s.app.Logger().Warn("No menu change detected", "name", r.Name)
		} else if scrapeErr != nil {
			s.app.Logger().Error("Error scraping restaurant", "name", r.Name, "error", scrapeErr)
		}

		s.markScrapeDone(restaurantID)
		status, detail := restaurant.LastCheckFromError(scrapeErr)
		if err := restaurant.UpdateLastCheck(s.app, restaurantID, status, detail); err != nil {
			s.app.Logger().Error("Failed to save last_check for restaurant", "id", restaurantID, "error", err)
		}
		s.notifyRestaurantStatus(restaurantID)
	}
}

func (s *Scraper) dequeueScrapeID() (string, bool) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	for len(s.scrapeQueue) == 0 && !s.queueClosed {
		s.queueCond.Wait()
	}

	if len(s.scrapeQueue) == 0 && s.queueClosed {
		return "", false
	}

	restaurantID := s.scrapeQueue[0]
	s.scrapeQueue = s.scrapeQueue[1:]
	delete(s.queued, restaurantID)
	s.inFlight[restaurantID] = struct{}{}
	s.lastRunAt[restaurantID] = time.Now()
	return restaurantID, true
}

func (s *Scraper) markScrapeDone(restaurantID string) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	delete(s.inFlight, restaurantID)
}

func (s *Scraper) StatusForRestaurant(restaurantID string) string {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	if _, ok := s.inFlight[restaurantID]; ok {
		return ScrapeStatusUpdating
	}
	if _, ok := s.queued[restaurantID]; ok {
		return ScrapeStatusQueued
	}
	if lastRunAt, ok := s.lastRunAt[restaurantID]; ok && time.Since(lastRunAt) < s.cooldown {
		return ScrapeStatusCooldown
	}

	return ScrapeStatusIdle
}

func (s *Scraper) notifyRestaurantStatus(restaurantID string) {
	status := s.StatusForRestaurant(restaurantID)

	payload := map[string]string{
		"id":     restaurantID,
		"status": status,
	}

	rawData, err := json.Marshal(payload)
	if err != nil {
		s.app.Logger().Error("Error marshaling restaurant status payload", "id", restaurantID, "error", err)
		return
	}

	message := subscriptions.Message{
		Name: RestaurantsStatusTopic,
		Data: rawData,
	}

	for _, chunk := range s.app.SubscriptionsBroker().ChunkedClients(300) {
		for _, client := range chunk {
			if !client.HasSubscription(RestaurantsStatusTopic) {
				continue
			}
			client.Send(message)
		}
	}
}

func (s *Scraper) getRestaurantByID(restaurantID string) (*restaurant.Restaurant, error) {
	return restaurant.GetRestaurant(s.app, restaurantID)
}

func (s *Scraper) scrapeSingle(r *restaurant.Restaurant) error {
	var err error

	initialDownloadPath := filepath.Join(restaurant.DownloadsFolder, fmt.Sprintf("%d_%s", time.Now().Unix(), r.ID))
	downloadPath := initialDownloadPath
	defer func() {
		if initialDownloadPath != downloadPath {
			_ = os.Remove(initialDownloadPath)
		}
		_ = os.Remove(downloadPath)
	}()

	switch r.Method {
	case "scrape":
		downloadPath, err = r.Scrape(downloadPath, s.web, s.app.Logger())
		if err != nil {
			return err
		}
	case "download":
		downloadPath, err = r.Download(downloadPath, s.app.Logger())
		if err != nil {
			return err
		}
	case "upload":
		return fmt.Errorf("%w: %s", restaurant.ErrManualUploadOnly, r.Name)
	default:
		return fmt.Errorf("unknown scraping method %q for restaurant %s", r.Method, r.Name)
	}

	return r.UpdateMenu(downloadPath, s.app)
}

func (s *Scraper) processFileToWebp(sourcePath string) (string, error) {
	tmpFilePath := filepath.Join(restaurant.DownloadsFolder, fmt.Sprintf("%d.webp", time.Now().UnixNano()))

	var err error
	if isPDFFile(sourcePath) {
		err = pdf.ConvertToWebp(sourcePath, tmpFilePath)
	} else {
		err = s.im.ConvertToWebp(sourcePath, tmpFilePath)
	}
	if err != nil {
		return "", err
	}

	if err = s.im.Trim(tmpFilePath); err != nil {
		os.Remove(tmpFilePath)
		return "", err
	}
	if err = s.im.ResizeWebp(tmpFilePath); err != nil {
		os.Remove(tmpFilePath)
		return "", err
	}

	return tmpFilePath, nil
}

func isPDFFile(sourcePath string) bool {
	if strings.EqualFold(filepath.Ext(sourcePath), ".pdf") {
		return true
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 512)
	readBytes, err := file.Read(header)
	if err != nil || readBytes == 0 {
		return false
	}

	return http.DetectContentType(header[:readBytes]) == "application/pdf"
}
