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
	DownloadsFolder = "data/downloads/"
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
			go m.scrapeMultiple(nil)
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
			m.scrapeMultiple(restaurants)
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

func (m *Mittag) scrapeMultiple(restaurants []*Restaurant) {
	var err error

	if restaurants == nil {
		restaurants, err = m.getRestaurants()
		if err != nil {
			m.app.Logger().Error("Error fetching restaurants", "error", err)
			return
		}
	}

	for _, r := range restaurants {
		if err = m.scrapeSingle(r); err != nil {
			m.app.Logger().Error("Error scraping restaurant", "id", r.ID, "error", err)
		}
	}
}

func (m *Mittag) scrapeSingle(restaurant *Restaurant) error {
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
		downloadPath, err = restaurant.Scrape(downloadPath, m.web, m.app.Logger())
		if err != nil {
			return err
		}
	case "download":
		downloadPath, err = restaurant.Download(downloadPath, m.app.Logger())
		if err != nil {
			return err
		}
	case "upload":
		m.app.Logger().Info("Restaurant is manually updated, skipping automated process", "id", restaurant.ID, "website", restaurant.Website)
		return nil
	default:
		m.app.Logger().Warn("Unknown scraping method for restaurant", "id", restaurant.ID, "method", restaurant.Method)
		return nil
	}

	tmpFilePath := filepath.Join(DownloadsFolder, fmt.Sprintf("%d_%s.webp", time.Now().Unix(), restaurant.ID))
	defer os.Remove(tmpFilePath)

	if restaurant.ContentType == "pdf" {
		err = pdf.ConvertToWebp(downloadPath, tmpFilePath)
	} else {
		err = m.im.ConvertToWebp(downloadPath, tmpFilePath)
	}
	if err != nil {
		return err
	}

	if err = m.im.Trim(tmpFilePath); err != nil {
		return err
	}

	if err = m.im.ResizeWebp(tmpFilePath); err != nil {
		return err
	}

	if err = restaurant.updateMenu(tmpFilePath, m.app); err != nil {
		return err
	}

	finalFilePath := filepath.Join(DownloadsFolder, fmt.Sprintf("%s.webp", restaurant.ID))
	if err = os.Rename(tmpFilePath, finalFilePath); err != nil {
		return err
	}

	return nil
}
