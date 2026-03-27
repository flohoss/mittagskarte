package mittag

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/internal/image"
	"github.com/flohoss/mittagskarte/internal/pdf"
	"github.com/flohoss/mittagskarte/internal/web"

	"github.com/pocketbase/pocketbase/core"
)

const (
	DownloadsFolder = "/app/data/downloads/"
)

func init() {
	os.MkdirAll(DownloadsFolder, os.ModePerm)
}

type Mittag struct {
	app         core.App
	restaurants []*Restaurant
	web         *web.Web
	im          *image.ImageMagic
}

func New(app core.App) (*Mittag, error) {
	webService, err := web.New()
	if err != nil {
		return nil, err
	}

	imageMagic := image.New()

	m := Mittag{app: app, web: webService, im: imageMagic}

	if err := m.initCron(); err != nil {
		return nil, err
	}

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/scrape", func(re *core.RequestEvent) error {
			go m.scrape(nil)
			return re.String(http.StatusOK, "Scraping started")
		})

		return se.Next()
	})

	return &m, nil
}

func (m *Mittag) Close() {
	m.web.Close()
	m.im.Close()
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
			m.scrape(restaurants)
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
		grouped[r.Cron] = append(grouped[r.Cron], restaurants[i])
	}

	return grouped, nil
}

func (m *Mittag) scrape(restaurants []*Restaurant) {
	if restaurants == nil {
		restaurants = m.restaurants
	}

	var err error

	for _, r := range restaurants {
		downloadPath := filepath.Join(DownloadsFolder, fmt.Sprintf("%d", time.Now().Unix()))
		defer os.Remove(downloadPath)

		switch r.Method {
		case "scrape":
			downloadPath, err = r.Scrape(downloadPath, m.web, m.app.Logger())
			if err != nil {
				m.app.Logger().Error("Error scraping restaurant", "id", r.ID, "error", err)
				continue
			}
		case "download":
			downloadPath, err = r.Download(downloadPath, m.app.Logger())
			if err != nil {
				m.app.Logger().Error("Error downloading menu", "id", r.ID, "error", err)
				continue
			}
		case "upload":
			m.app.Logger().Info("Restaurant is manually updated, skipping automated process", "id", r.ID, "website", r.Website)
		default:
			m.app.Logger().Warn("Unknown scraping method for restaurant", "id", r.ID, "method", r.Method)
		}

		filePath := filepath.Join(DownloadsFolder, r.ID+".webp")

		if r.ContentType == "pdf" {
			err = pdf.ConvertToWebp(downloadPath, filePath)
		} else {
			err = m.im.ConvertToWebp(downloadPath, filePath)
		}
		if err != nil {
			m.app.Logger().Error("Error converting menu to webp", "id", r.ID, "error", err)
			continue
		}

		if err = m.im.Trim(filePath); err != nil {
			m.app.Logger().Error("Error trimming menu image", "id", r.ID, "error", err)
			continue
		}

		if err = m.im.ResizeWebp(filePath); err != nil {
			m.app.Logger().Error("Error resizing menu image", "id", r.ID, "error", err)
			continue
		}

		if err = r.updateMenu(filePath, m.app); err != nil {
			m.app.Logger().Error("Error updating restaurant menu", "id", r.ID, "error", err)
			continue
		}

		m.app.Logger().Info("Successfully updated menu for restaurant", "id", r.ID, "filePath", filePath)
	}
}
