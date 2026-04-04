package mittag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/internal/image"
	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/flohoss/mittagskarte/internal/web"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type Mittag struct {
	app     core.App
	scraper *Scraper
	started bool
}

func New(app core.App) (*Mittag, error) {
	webService, err := web.New()
	if err != nil {
		return nil, err
	}

	imageMagic := image.New()

	m := &Mittag{app: app}
	m.scraper = NewScraper(app, webService, imageMagic, restaurant.FetchRestaurants)
	m.bindHooks()

	return m, nil
}

func (m *Mittag) Start() error {
	if m.started {
		return nil
	}

	if err := m.initCron(); err != nil {
		return err
	}

	m.started = true
	return nil
}

func (m *Mittag) Close() {
	m.scraper.Close()
}

func (m *Mittag) initCron() error {
	crons, err := restaurant.FetchCronGroups(m.app)
	if err != nil {
		return err
	}

	m.app.Logger().Debug("Initializing cron jobs for restaurant groups", "groups", len(crons))

	for cron, restaurants := range crons {
		restaurantIDs := make([]string, len(restaurants))
		for i, r := range restaurants {
			restaurantIDs[i] = r.ID
		}
		m.app.Logger().Debug("Adding cron for restaurant group", "cron", cron, "restaurants", strings.Join(restaurantIDs, ","))
		m.app.Cron().MustAdd(cron, cron, func() {
			m.scraper.Enqueue(restaurants)
		})
	}

	return nil
}

func (m *Mittag) bindHooks() {
	m.app.OnRecordEnrich("restaurants").BindFunc(func(e *core.RecordEnrichEvent) error {
		e.Record.WithCustomData(true)
		e.Record.Set("status", m.scraper.StatusForRestaurant(e.Record.Id))
		return e.Next()
	})

	m.app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/restaurants/scrape", m.handleScrape).Bind(apis.RequireAuth())
		se.Router.POST("/api/restaurants/upload", m.handleUpload).Bind(apis.RequireAuth())

		fileServer := http.FileServer(http.Dir(restaurant.DownloadsFolder))
		se.Router.GET("/data/downloads/{path...}", func(re *core.RequestEvent) error {
			http.StripPrefix("/data/downloads/", fileServer).ServeHTTP(re.Response, re.Request)
			return nil
		})

		return se.Next()
	})
}

func (m *Mittag) handleScrape(re *core.RequestEvent) error {
	var payload struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(re.Request.Body).Decode(&payload); err != nil {
		return re.String(http.StatusBadRequest, "Invalid JSON body")
	}

	r, err := m.loadRestaurant(re, payload.ID)
	if err != nil {
		return err
	}

	m.scraper.Enqueue([]*restaurant.Restaurant{r})

	return re.String(http.StatusOK, fmt.Sprintf("Scrape triggered for restaurant %s", r.ID))
}

func (m *Mittag) handleUpload(re *core.RequestEvent) error {
	re.Request.Body = http.MaxBytesReader(re.Response, re.Request.Body, 25<<20)
	if err := re.Request.ParseMultipartForm(25 << 20); err != nil {
		return re.String(http.StatusBadRequest, "Invalid multipart form data")
	}

	r, err := m.loadRestaurant(re, re.Request.FormValue("id"))
	if err != nil {
		return err
	}
	if r.Method != "upload" {
		return re.String(http.StatusBadRequest, "Restaurant does not support uploads")
	}

	file, header, err := re.Request.FormFile("file")
	if err != nil {
		return re.String(http.StatusBadRequest, "file is required")
	}
	defer file.Close()

	uploadPath := filepath.Join(restaurant.DownloadsFolder, fmt.Sprintf("upload_%d_%s_%s", time.Now().UnixNano(), r.ID, filepath.Base(header.Filename)))
	out, err := os.Create(uploadPath)
	if err != nil {
		return re.String(http.StatusInternalServerError, "Could not create upload file")
	}
	_, copyErr := io.Copy(out, file)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(uploadPath)
		return re.String(http.StatusInternalServerError, "Could not persist uploaded file")
	}
	defer os.Remove(uploadPath)

	if err := m.scraper.uploadSingle(r, uploadPath); err != nil {
		m.app.Logger().Error("Error uploading restaurant menu", "id", r.ID, "error", err)
		return re.String(http.StatusInternalServerError, "Could not process uploaded file")
	}

	return re.String(http.StatusOK, fmt.Sprintf("Upload processed for restaurant %s", r.ID))
}

func (m *Mittag) loadRestaurant(re *core.RequestEvent, id string) (*restaurant.Restaurant, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, re.String(http.StatusBadRequest, "restaurant id is required")
	}
	r, err := restaurant.FetchRestaurant(m.app, id)
	if err != nil {
		return nil, re.String(http.StatusInternalServerError, "Could not load restaurant")
	}
	return r, nil
}
