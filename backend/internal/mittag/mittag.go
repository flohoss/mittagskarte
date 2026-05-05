package mittag

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/internal/image"
	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/flohoss/mittagskarte/internal/web"
	"github.com/flohoss/mittagskarte/pkg/checksum"
	"github.com/flohoss/mittagskarte/pkg/fsutil"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/router"
)

type Mittag struct {
	app     core.App
	scraper *Scraper
	started bool
}

func New(app core.App, coolDownDuration time.Duration) (*Mittag, error) {
	webService, err := web.New()
	if err != nil {
		return nil, err
	}

	imageMagic := image.New()

	m := &Mittag{app: app}
	m.scraper = NewScraper(app, webService, imageMagic, restaurant.GetRestaurantsWithNavigate, coolDownDuration)
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
	crons, err := restaurant.GetCronGroups(m.app)
	if err != nil {
		return err
	}

	m.app.Logger().Debug("Initializing cron jobs for restaurant groups", "groups", len(crons))

	for cron, restaurants := range crons {
		names := make([]string, len(restaurants))
		for i, r := range restaurants {
			names[i] = r.Name
		}
		m.app.Logger().Debug("Adding cron for restaurant group", "cron", cron, "restaurants", strings.Join(names, ","))
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

	m.app.OnRecordCreate("menus").BindFunc(m.onMenuCreate)
	m.app.OnRecordAfterCreateSuccess("menus").BindFunc(m.onMenuAfterCreateSuccess)

	m.app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/restaurants/scrape", m.handleScrape).Bind(apis.RequireAuth())
		return se.Next()
	})
}

func (m *Mittag) onMenuCreate(e *core.RecordEvent) error {
	restaurantID := e.Record.GetString("restaurant")
	if restaurantID == "" {
		return e.Next()
	}

	f, ok := e.Record.Get("file").(*filesystem.File)
	if !ok || f == nil {
		return e.Next()
	}

	sourcePath, cleanup, err := fsutil.LocalPath(f, restaurant.DownloadsFolder)
	if err != nil {
		return router.NewBadRequestError("Hochgeladene Datei konnte nicht verarbeitet werden", err)
	}
	defer cleanup()

	tmpWebp, err := m.scraper.processFileToWebp(sourcePath)
	if err != nil {
		return router.NewBadRequestError("Menü konnte nicht verarbeitet werden", err)
	}
	defer os.Remove(tmpWebp)

	// Set dimensions while tmpWebp still exists on disk.
	restaurant.SetMenuDimensions(e.Record, tmpWebp)

	// Read into memory before removing the temp file.
	// PocketBase stores the file after the hook returns, so a PathReader
	// pointing to a deleted temp file would fail — BytesReader is safe.
	data, readErr := os.ReadFile(tmpWebp)
	if readErr != nil {
		return router.NewBadRequestError("Verarbeitetes Menü konnte nicht gelesen werden", readErr)
	}

	processedFile := &filesystem.File{
		Reader: &filesystem.BytesReader{Bytes: data},
		Name:   "menu.webp",
		Size:   int64(len(data)),
	}
	e.Record.Set("file", processedFile)

	rc, err := processedFile.Reader.Open()
	if err != nil {
		return e.Next()
	}
	hash, err := checksum.Reader(rc)
	rc.Close()
	if err != nil {
		return e.Next()
	}

	if latest := restaurant.GetLatestMenuByRestaurantID(m.app, restaurantID); latest != nil {
		if latest.GetString("hash") == hash {
			m.app.Logger().Debug("Menu has not changed, skipping update", "restaurantId", restaurantID)
			status, detail := restaurant.LastCheckFromError(restaurant.ErrMenuUnchanged)
			if err := restaurant.UpdateLastCheck(m.app, restaurantID, status, detail); err != nil {
				m.app.Logger().Error("Failed to update last_check for unchanged menu", "restaurantId", restaurantID, "error", err)
			}
			return router.NewBadRequestError("Das Menü hat sich nicht geändert", fmt.Errorf("%w", restaurant.ErrMenuUnchanged))
		}
	}

	e.Record.Set("hash", hash)
	return e.Next()
}

func (m *Mittag) onMenuAfterCreateSuccess(e *core.RecordEvent) error {
	restaurantID := e.Record.GetString("restaurant")
	if restaurantID == "" {
		return e.Next()
	}

	restaurantRecord, err := m.app.FindRecordById("restaurants", restaurantID)
	if err != nil {
		m.app.Logger().Error("Failed to find restaurant for menu", "restaurantId", restaurantID, "error", err)
		return e.Next()
	}

	retentionLimit, err := m.menuRetentionLimit()
	if err != nil {
		m.app.Logger().Error("Failed to resolve menu retention limit", "restaurantId", restaurantID, "error", err)
		return e.Next()
	}

	// Prepend new menu to the existing relation IDs (already newest-first).
	relationIDs := append([]string{e.Record.Id}, restaurantRecord.GetStringSlice("menus")...)

	if len(relationIDs) > retentionLimit {
		for _, oldID := range relationIDs[retentionLimit:] {
			if old, findErr := m.app.FindRecordById("menus", oldID); findErr == nil {
				if deleteErr := m.app.Delete(old); deleteErr != nil {
					m.app.Logger().Warn("Failed to delete old menu during retention cleanup", "restaurantId", restaurantID, "menuId", oldID, "error", deleteErr)
				}
			}
		}
		relationIDs = relationIDs[:retentionLimit]
	}

	restaurantRecord.Set("menus", relationIDs)
	status, detail := restaurant.LastCheckFromError(nil)
	restaurant.SetLastCheck(restaurantRecord, status, detail)
	if err := m.app.Save(restaurantRecord); err != nil {
		m.app.Logger().Error("Failed to update restaurant menus", "restaurantId", restaurantID, "error", err)
	}

	return e.Next()
}

func (m *Mittag) handleScrape(re *core.RequestEvent) error {
	var payload struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(re.Request.Body).Decode(&payload); err != nil {
		return re.String(http.StatusBadRequest, "Ungültiger JSON-Body")
	}

	r, err := m.loadRestaurant(re, payload.ID)
	if err != nil {
		return err
	}

	m.scraper.Enqueue([]*restaurant.Restaurant{r})

	return re.String(http.StatusOK, fmt.Sprintf("Aktualisierung für Restaurant %s gestartet", r.ID))
}

func (m *Mittag) loadRestaurant(re *core.RequestEvent, id string) (*restaurant.Restaurant, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, re.String(http.StatusBadRequest, "Restaurant-ID ist erforderlich")
	}
	r, err := restaurant.GetRestaurant(m.app, id)
	if err != nil {
		return nil, re.String(http.StatusInternalServerError, "Restaurant konnte nicht geladen werden")
	}
	return r, nil
}

func (m *Mittag) menuRetentionLimit() (int, error) {
	restaurantsCollection, err := m.app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		return 0, err
	}

	menusField, ok := restaurantsCollection.Fields.GetByName("menus").(*core.RelationField)
	if !ok {
		return 0, fmt.Errorf("restaurants.menus is not a relation field")
	}

	// PocketBase relation fields are single-select when MaxSelect <= 1.
	if menusField.MaxSelect <= 1 {
		return 1, nil
	}

	return menusField.MaxSelect, nil
}
