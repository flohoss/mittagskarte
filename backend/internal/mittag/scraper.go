package mittag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/flohoss/mittagskarte/internal/web"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/subscriptions"
)

const (
	ScrapeStatusUpdating = "updating"
	ScrapeStatusQueued   = "queued"
	ScrapeStatusCoolDown = "cooldown"
	ScrapeStatusIdle     = "idle"

	RestaurantsStatusTopic = "restaurants/status"
)

type queueItem struct {
	restaurantID     string
	onDone           func(err error, r *restaurant.Restaurant)
	telegramFilePath string
}

type Scraper struct {
	app core.App
	web *web.Web

	queueMu     sync.Mutex
	queueCond   *sync.Cond
	scrapeQueue []queueItem
	queued      map[string]struct{}
	inFlight    map[string]struct{}
	lastRunAt   map[string]time.Time
	coolDown    time.Duration
	queueClosed bool
	workerWg    sync.WaitGroup
}

func NewScraper(app core.App, webService *web.Web, coolDownDuration time.Duration) *Scraper {
	s := &Scraper{
		app:       app,
		web:       webService,
		queued:    make(map[string]struct{}),
		inFlight:  make(map[string]struct{}),
		lastRunAt: make(map[string]time.Time),
		coolDown:  coolDownDuration,
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

	if s.web != nil {
		s.web.Close()
	}
}

func (s *Scraper) Enqueue(restaurants []*restaurant.Restaurant) {
	queuedRestaurantIDs := s.enqueueItemsLocked(s.itemsFromRestaurants(restaurants))

	for _, restaurantID := range queuedRestaurantIDs {
		s.notifyRestaurantStatus(restaurantID)
	}
}

func (s *Scraper) itemsFromRestaurants(restaurants []*restaurant.Restaurant) []queueItem {
	items := make([]queueItem, 0, len(restaurants))
	for _, r := range restaurants {
		items = append(items, queueItem{restaurantID: r.ID})
	}
	return items
}

func (s *Scraper) EnqueueTelegramUpload(r *restaurant.Restaurant, filePath string, onDone func(error, *restaurant.Restaurant)) {
	item := queueItem{
		restaurantID:     r.ID,
		onDone:           onDone,
		telegramFilePath: filePath,
	}
	queuedRestaurantIDs := s.enqueueItemsLocked([]queueItem{item})

	for _, restaurantID := range queuedRestaurantIDs {
		s.notifyRestaurantStatus(restaurantID)
	}
}

func (s *Scraper) menuItems(restaurants []*restaurant.Restaurant) map[string]queueItem {
	items := make(map[string]queueItem, len(restaurants))
	for _, r := range restaurants {
		items[r.ID] = queueItem{restaurantID: r.ID}
	}
	return items
}

func (s *Scraper) enqueueItemsLocked(items []queueItem) []string {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	if s.queueClosed {
		s.app.Logger().Warn("Scrape queue is closed, skipping enqueue")
		return nil
	}

	now := time.Now()
	queuedRestaurantIDs := make([]string, 0)
	for _, item := range items {
		if _, ok := s.queued[item.restaurantID]; ok {
			s.app.Logger().Debug("Skipping enqueue: restaurant already queued", "id", item.restaurantID)
			continue
		}
		if _, ok := s.inFlight[item.restaurantID]; ok {
			s.app.Logger().Debug("Skipping enqueue: restaurant scrape in progress", "id", item.restaurantID)
			continue
		}
		if lastRunAt, ok := s.lastRunAt[item.restaurantID]; ok && now.Sub(lastRunAt) < s.coolDown {
			s.app.Logger().Debug("Skipping enqueue: restaurant in coolDown", "id", item.restaurantID, "nextAllowedAt", lastRunAt.Add(s.coolDown))
			continue
		}

		s.scrapeQueue = append(s.scrapeQueue, item)
		s.queued[item.restaurantID] = struct{}{}
		queuedRestaurantIDs = append(queuedRestaurantIDs, item.restaurantID)
	}

	s.queueCond.Broadcast()

	return queuedRestaurantIDs
}

func (s *Scraper) runScrapeWorker() {
	defer s.workerWg.Done()

	for {
		item, ok := s.dequeueScrape()
		if !ok {
			return
		}

		s.processScrape(item)
	}
}

func (s *Scraper) processScrape(item queueItem) {
	defer func() {
		if r := recover(); r != nil {
			s.app.Logger().Error("Panic during scrape, recovering", "id", item.restaurantID, "panic", r)
			s.markScrapeDone(item.restaurantID)
			s.notifyRestaurantStatus(item.restaurantID)
		}
	}()

	s.notifyRestaurantStatus(item.restaurantID)

	r, err := s.getRestaurantByID(item.restaurantID)
	if err != nil {
		s.app.Logger().Error("Error resolving restaurant for scrape", "id", item.restaurantID, "error", err)
		s.markScrapeDone(item.restaurantID)
		s.notifyRestaurantStatus(item.restaurantID)
		return
	}

	scrapeErr := s.scrapeSingle(item, r)
	if scrapeErr == nil || errors.Is(scrapeErr, restaurant.ErrManualUploadOnly) || errors.Is(scrapeErr, restaurant.ErrMenuUnchanged) {
		s.app.Logger().Debug("Scrape finished without change", "id", item.restaurantID, "name", r.Name)
	} else {
		s.app.Logger().Error("Error scraping restaurant", "id", item.restaurantID, "name", r.Name, "error", scrapeErr)
	}
	s.finishScrape(item, r, scrapeErr)
}

func (s *Scraper) finishScrape(item queueItem, r *restaurant.Restaurant, scrapeErr error) {
	s.markScrapeDone(item.restaurantID)
	status, detail := restaurant.LastCheckFromError(scrapeErr)
	if err := restaurant.UpdateLastCheck(s.app, item.restaurantID, status, detail); err != nil {
		s.app.Logger().Error("Failed to save last_check for restaurant", "id", item.restaurantID, "error", err)
	}
	s.notifyRestaurantStatus(item.restaurantID)

	if item.onDone != nil {
		item.onDone(scrapeErr, r)
	}
}

func (s *Scraper) dequeueScrape() (queueItem, bool) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	for len(s.scrapeQueue) == 0 && !s.queueClosed {
		s.queueCond.Wait()
	}

	if len(s.scrapeQueue) == 0 && s.queueClosed {
		return queueItem{}, false
	}

	item := s.scrapeQueue[0]
	s.scrapeQueue = s.scrapeQueue[1:]
	delete(s.queued, item.restaurantID)
	s.inFlight[item.restaurantID] = struct{}{}
	s.lastRunAt[item.restaurantID] = time.Now()
	return item, true
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
	if lastRunAt, ok := s.lastRunAt[restaurantID]; ok && time.Since(lastRunAt) < s.coolDown {
		return ScrapeStatusCoolDown
	}

	return ScrapeStatusIdle
}

func (s *Scraper) notifyRestaurantStatus(restaurantID string) {
	status := s.StatusForRestaurant(restaurantID)

	coolDownSeconds := 0
	s.queueMu.Lock()
	if lastRunAt, ok := s.lastRunAt[restaurantID]; ok && time.Since(lastRunAt) < s.coolDown {
		coolDownSeconds = int(s.coolDown.Seconds() - time.Since(lastRunAt).Seconds())
		if coolDownSeconds < 0 {
			coolDownSeconds = 0
		}
	}
	s.queueMu.Unlock()

	payload := map[string]interface{}{
		"id":              restaurantID,
		"status":          status,
		"coolDownSeconds": coolDownSeconds,
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

func (s *Scraper) scrapeSingle(item queueItem, r *restaurant.Restaurant) error {
	var err error

	if item.telegramFilePath != "" {
		s.app.Logger().Debug("Storing Telegram upload as menu", "id", r.ID, "name", r.Name, "path", item.telegramFilePath)
		return r.UpdateMenu(item.telegramFilePath, s.app)
	}

	initialDownloadPath := filepath.Join(restaurant.DownloadsFolder, fmt.Sprintf("%d_%s", time.Now().Unix(), r.ID))
	downloadPath := initialDownloadPath
	defer func() {
		if initialDownloadPath != downloadPath {
			_ = os.Remove(initialDownloadPath)
		}
		_ = os.Remove(downloadPath)
	}()

	s.app.Logger().Debug("Starting scrape", "id", r.ID, "name", r.Name, "method", r.Method)

	switch r.Method {
	case restaurant.MethodScrape:
		downloadPath, err = r.Scrape(downloadPath, s.web, s.app.Logger())
		if err != nil {
			return err
		}
	case restaurant.MethodDownload:
		downloadPath, err = r.Download(downloadPath, s.app.Logger())
		if err != nil {
			return err
		}
	case restaurant.MethodUpload:
		return fmt.Errorf("%w: %s", restaurant.ErrManualUploadOnly, r.Name)
	default:
		return fmt.Errorf("unknown scraping method %q for restaurant %s", r.Method, r.Name)
	}

	s.app.Logger().Debug("Scrape completed, updating menu", "id", r.ID, "name", r.Name, "downloadPath", downloadPath)
	return r.UpdateMenu(downloadPath, s.app)
}
