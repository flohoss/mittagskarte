package mittag

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/flohoss/mittagskarte/internal/snapotter"
	"github.com/flohoss/mittagskarte/internal/web"
	"github.com/flohoss/mittagskarte/pkg/checksum"
	"github.com/flohoss/mittagskarte/pkg/fsutil"
	"github.com/flohoss/mittagskarte/pkg/pdfinfo"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/router"
)

type Mittag struct {
	app       core.App
	scraper   *Scraper
	snapotter *snapotter.Client
	started   bool
}

func New(app core.App, snapOtterURL url.URL, coolDownDuration time.Duration) (*Mittag, error) {
	webService, err := web.New()
	if err != nil {
		return nil, err
	}

	snapOtterClient := snapotter.New(snapOtterURL, app.Logger())
	if err := snapOtterClient.Setup(); err != nil {
		return nil, fmt.Errorf("setup snapotter: %w", err)
	}

	m := &Mittag{app: app, snapotter: snapOtterClient}
	m.scraper = NewScraper(app, webService, restaurant.GetRestaurantsWithNavigate, coolDownDuration)
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
		m.app.Logger().Error("Hochgeladene Datei konnte nicht verarbeitet werden", "restaurantId", restaurantID, "error", err)
		return router.NewBadRequestError("Hochgeladene Datei konnte nicht verarbeitet werden", err)
	}
	defer cleanup()

	m.app.Logger().Debug("Processing menu file", "restaurantId", restaurantID, "sourcePath", sourcePath)

	if pdfinfo.IsPDF(sourcePath) {
		pdfInfo, pdfErr := pdfinfo.Read(sourcePath)
		if pdfErr != nil {
			m.app.Logger().Warn("Failed to inspect PDF metadata, continuing with conversion", "restaurantId", restaurantID, "error", pdfErr)
		} else {
			e.Record.Set("pdf_metadata", pdfInfo)
			if m.pdfUnchanged(restaurantID, pdfInfo) {
				m.app.Logger().Debug("PDF metadata unchanged, skipping conversion", "restaurantId", restaurantID)
				restaurant.UpdateLastCheck(m.app, restaurantID, restaurant.LastCheckStatusNotChanged, "")
				return router.NewBadRequestError("Das Menü hat sich nicht geändert", fmt.Errorf("%w", restaurant.ErrMenuUnchanged))
			}
		}
	}

	result, err := m.snapotter.ProcessFileToWebp(sourcePath)
	if err != nil {
		m.app.Logger().Error("Menü konnte nicht verarbeitet werden", "restaurantId", restaurantID, "sourcePath", sourcePath, "error", err)
		return router.NewBadRequestError("Menü konnte nicht verarbeitet werden", err)
	}

	m.app.Logger().Debug("Menu file processed", "restaurantId", restaurantID, "width", result.Width, "height", result.Height, "bytes", len(result.Data))

	e.Record.Set("dimensions", map[string]any{
		"width":     result.Width,
		"height":    result.Height,
		"landscape": result.Width >= result.Height,
	})

	processedFile := &filesystem.File{
		Reader: &filesystem.BytesReader{Bytes: result.Data},
		Name:   "menu.webp",
		Size:   int64(len(result.Data)),
	}
	e.Record.Set("file", processedFile)

	rc, err := processedFile.Reader.Open()
	if err != nil {
		m.app.Logger().Error("Failed to open processed menu file for checksum", "restaurantId", restaurantID, "error", err)
		return e.Next()
	}
	hash, err := checksum.Reader(rc)
	rc.Close()
	if err != nil {
		m.app.Logger().Error("Failed to compute menu checksum", "restaurantId", restaurantID, "error", err)
		return e.Next()
	}

	m.app.Logger().Debug("Computed menu checksum", "restaurantId", restaurantID, "hash", hash)

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

func (m *Mittag) pdfUnchanged(restaurantID string, current pdfinfo.Metadata) bool {
	latest := restaurant.GetLatestMenuByRestaurantID(m.app, restaurantID)
	if latest == nil {
		return false
	}
	return pdfinfo.Equal(latest.Get("pdf_metadata"), current)
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

	menuRecords, err := m.app.FindRecordsByFilter("menus", "restaurant = {:id}", "-created", 0, 0, dbx.Params{"id": restaurantID})
	if err != nil {
		m.app.Logger().Error("Failed to list menus for retention cleanup", "restaurantId", restaurantID, "error", err)
		return e.Next()
	}

	relationIDs := make([]string, 0, retentionLimit)
	if len(menuRecords) > 0 {
		for i, record := range menuRecords {
			if i < retentionLimit {
				relationIDs = append(relationIDs, record.Id)
				continue
			}

			if deleteErr := m.app.Delete(record); deleteErr != nil {
				m.app.Logger().Warn("Failed to delete old menu during retention cleanup", "restaurantId", restaurantID, "menuId", record.Id, "error", deleteErr)
			}
		}
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
