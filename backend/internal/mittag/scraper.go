package mittag

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flohoss/mittagskarte/internal/image"
	"github.com/flohoss/mittagskarte/internal/pdf"
	"github.com/flohoss/mittagskarte/internal/web"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
)

type restaurantsProvider func() ([]*Restaurant, error)

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
		cooldown:       10 * time.Minute,
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

func (s *Scraper) Enqueue(restaurants []*Restaurant) {
	var err error

	if restaurants == nil {
		restaurants, err = s.getRestaurants()
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

		restaurant, err := s.getRestaurantByID(restaurantID)
		if err != nil {
			s.app.Logger().Error("Error resolving restaurant for scrape", "id", restaurantID, "error", err)
			continue
		}

		if err = s.scrapeSingle(restaurant); err != nil {
			s.app.Logger().Error("Error scraping restaurant", "id", restaurant.ID, "error", err)
		}

		s.markScrapeDone(restaurantID)
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

func (s *Scraper) getRestaurantByID(restaurantID string) (*Restaurant, error) {
	restaurants, err := s.getRestaurants()
	if err != nil {
		return nil, err
	}

	for _, restaurant := range restaurants {
		if restaurant.ID == restaurantID {
			return restaurant, nil
		}
	}

	return nil, fmt.Errorf("restaurant %s not found", restaurantID)
}

func (s *Scraper) uploadSingle(restaurant *Restaurant, uploadPath string) error {
	if restaurant == nil {
		return fmt.Errorf("restaurant is nil")
	}

	return s.processAndUpdateMenu(restaurant, uploadPath)
}

func (s *Scraper) scrapeSingle(restaurant *Restaurant) error {
	var err error

	initialDownloadPath := filepath.Join(DownloadsFolder, fmt.Sprintf("%d_%s", time.Now().Unix(), restaurant.ID))
	downloadPath := initialDownloadPath
	defer func() {
		if initialDownloadPath != downloadPath {
			_ = os.Remove(initialDownloadPath)
		}
		_ = os.Remove(downloadPath)
	}()

	switch restaurant.Method {
	case "scrape":
		downloadPath, err = restaurant.Scrape(downloadPath, s.web, s.app.Logger())
		if err != nil {
			return err
		}
	case "download":
		downloadPath, err = restaurant.Download(downloadPath, s.app.Logger())
		if err != nil {
			return err
		}
	case "upload":
		s.app.Logger().Info("Restaurant is manually updated, skipping automated process", "id", restaurant.ID, "website", restaurant.Website)
		return nil
	default:
		s.app.Logger().Warn("Unknown scraping method for restaurant", "id", restaurant.ID, "method", restaurant.Method)
		return nil
	}

	return s.processAndUpdateMenu(restaurant, downloadPath)
}

func (s *Scraper) processAndUpdateMenu(restaurant *Restaurant, sourcePath string) error {
	tmpFilePath := filepath.Join(DownloadsFolder, fmt.Sprintf("%d_%s.webp", time.Now().UnixNano(), restaurant.ID))
	defer os.Remove(tmpFilePath)

	var err error
	if shouldUsePDFConverter(restaurant, sourcePath) {
		err = pdf.ConvertToWebp(sourcePath, tmpFilePath)
	} else {
		err = s.im.ConvertToWebp(sourcePath, tmpFilePath)
	}
	if err != nil {
		return err
	}

	if err = s.im.Trim(tmpFilePath); err != nil {
		return err
	}

	if err = s.im.ResizeWebp(tmpFilePath); err != nil {
		return err
	}

	finalFilePath := filepath.Join(DownloadsFolder, fmt.Sprintf("%s.webp", restaurant.ID))
	if err = os.Rename(tmpFilePath, finalFilePath); err != nil {
		return err
	}

	if err = restaurant.updateMenu(finalFilePath, s.app); err != nil {
		return err
	}

	return nil
}

func shouldUsePDFConverter(restaurant *Restaurant, sourcePath string) bool {
	if restaurant.ContentType == "pdf" {
		return true
	}

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
