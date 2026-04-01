package mittag

import (
	"net/http"
	"os"
	"strings"

	"github.com/flohoss/mittagskarte/internal/image"
	"github.com/flohoss/mittagskarte/internal/web"

	"github.com/pocketbase/pocketbase/core"
)

const (
	DownloadsFolder = "data/downloads/"
)

func init() {
	os.MkdirAll(DownloadsFolder, os.ModePerm)
}

type Mittag struct {
	app         core.App
	domain      string
	restaurants []*Restaurant
	scraper     *Scraper
}

func New(app core.App, domain string) (*Mittag, error) {
	webService, err := web.New()
	if err != nil {
		return nil, err
	}

	imageMagic := image.New()

	m := &Mittag{app: app, domain: domain}
	m.scraper = NewScraper(app, webService, imageMagic, m.getRestaurants)
	m.bindHooks()

	if err := m.initCron(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Mittag) Close() {
	m.scraper.Close()
}

func (m *Mittag) initCron() error {
	crons, err := m.getCronGroups()
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

func (m *Mittag) getRestaurants() ([]*Restaurant, error) {
	if m.restaurants != nil {
		return m.restaurants, nil
	}
	restaurants, err := fetchRestaurants(m.app)
	if err != nil {
		return nil, err
	}
	m.restaurants = restaurants
	return restaurants, nil
}

func (m *Mittag) getCronGroups() (map[string][]*Restaurant, error) {
	restaurants, err := m.getRestaurants()
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]*Restaurant)
	for i, r := range restaurants {
		if r.Cron == "" {
			continue
		}
		grouped[r.Cron] = append(grouped[r.Cron], restaurants[i])
	}

	return grouped, nil
}

func (m *Mittag) bindHooks() {
	m.app.OnRecordEnrich("restaurants").BindFunc(func(e *core.RecordEnrichEvent) error {
		e.Record.WithCustomData(true)
		e.Record.Set("scrape_status", m.scraper.StatusForRestaurant(e.Record.Id))
		return e.Next()
	})

	m.app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/scrape", func(re *core.RequestEvent) error {
			go m.scraper.Enqueue(nil)
			return re.String(http.StatusOK, "Scraping started")
		})

		fileServer := http.FileServer(http.Dir(DownloadsFolder))
		se.Router.GET("/data/downloads/{path...}", func(re *core.RequestEvent) error {
			http.StripPrefix("/data/downloads/", fileServer).ServeHTTP(re.Response, re.Request)
			return nil
		})

		return se.Next()
	})
}
