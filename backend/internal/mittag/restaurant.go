package mittag

import (
	"fmt"
	"image"
	"log/slog"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/flohoss/mittagskarte/internal/checksum"
	"github.com/flohoss/mittagskarte/internal/download"
	"github.com/flohoss/mittagskarte/internal/placeholder"
	"github.com/flohoss/mittagskarte/internal/web"

	"github.com/playwright-community/playwright-go"
	"github.com/pocketbase/pocketbase/core"
	_ "golang.org/x/image/webp"
)

type menuDimensions struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Landscape bool `json:"landscape"`
}

type Restaurant struct {
	ID          string     `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Website     string     `db:"website" json:"website"`
	RestDays    []string   `db:"rest_days" json:"rest_days"`
	Method      string     `db:"method" json:"method"`
	ContentType string     `db:"content_type" json:"content_type"`
	Cron        string     `db:"cron" json:"cron"`
	Navigate    []Selector `db:"navigate" json:"navigate"`
}

type Selector struct {
	Id        string `db:"id" json:"id"`
	Order     int    `db:"order" json:"order"`
	Locator   string `db:"locator" json:"locator"`
	Attribute string `db:"attribute" json:"attribute"`
	Style     string `db:"style" json:"style"`
}

func (r *Restaurant) withLogger(logger *slog.Logger, extra ...any) *slog.Logger {
	args := make([]any, 0, len(extra)+2)
	args = append(args, "name", r.Name)
	args = append(args, extra...)
	return logger.With(args...)
}

func fetchRestaurants(app core.App) ([]*Restaurant, error) {
	r, err := app.FindAllRecords("restaurants")
	if err != nil {
		return nil, err
	}
	app.ExpandRecords(r, []string{"navigate"}, nil)

	restaurants := make([]*Restaurant, len(r))
	for i, restaurant := range r {
		expandedNavigate := restaurant.ExpandedAll("navigate")
		navigate := make([]Selector, 0, len(expandedNavigate))
		for _, nav := range expandedNavigate {
			navigate = append(navigate, Selector{
				Id:        nav.Id,
				Order:     nav.GetInt("order"),
				Locator:   nav.GetString("locator"),
				Attribute: nav.GetString("attribute"),
				Style:     nav.GetString("style"),
			})
		}

		sort.SliceStable(navigate, func(i, j int) bool {
			return navigate[i].Order < navigate[j].Order
		})

		restaurants[i] = &Restaurant{
			ID:          restaurant.Id,
			Name:        restaurant.GetString("name"),
			Website:     restaurant.GetString("website"),
			RestDays:    restaurant.GetStringSlice("rest_days"),
			Method:      restaurant.GetString("method"),
			ContentType: restaurant.GetString("content_type"),
			Cron:        restaurant.GetString("cron"),
			Navigate:    navigate,
		}
	}

	return restaurants, nil
}

func (r *Restaurant) updateMenu(filePath string, app core.App) error {
	logger := r.withLogger(app.Logger())

	restaurant, err := app.FindRecordById("restaurants", r.ID)
	if err != nil {
		return err
	}

	existingChecksum := restaurant.GetString("menu_hash")
	newChecksum, err := checksum.ChecksumFile(filePath)
	if err != nil {
		return err
	}

	if checksum.Identical(existingChecksum, newChecksum) {
		logger.Info("Menu has not changed, skipping update")
		return nil
	}

	filePathWithChecksum, err := checksum.SuffixQuery(filePath)
	if err != nil {
		return err
	}

	dimensions, err := readMenuDimensions(filePath)
	if err != nil {
		logger.Warn("Could not read menu image dimensions", "path", filePath, "error", err)
	}

	restaurant.Set("menu", filePathWithChecksum)
	restaurant.Set("menu_hash", fmt.Sprintf("%x", newChecksum))
	if dimensions != nil {
		restaurant.Set("menu_dimensions", dimensions)
	}
	if err := app.Save(restaurant); err != nil {
		return err
	}

	logger.Info("Successfully updated menu for restaurant", "filePath", filePathWithChecksum)

	return nil
}

func readMenuDimensions(filePath string) (*menuDimensions, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	if config.Width <= 0 || config.Height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", config.Width, config.Height)
	}

	return &menuDimensions{
		Width:     config.Width,
		Height:    config.Height,
		Landscape: config.Width >= config.Height,
	}, nil
}

func (r *Restaurant) Download(downloadPath string, logger *slog.Logger) (string, error) {
	logger = r.withLogger(logger, "website", r.Website)
	logger.Info("Downloading menu from direct URL")

	if len(r.Navigate) == 0 || r.Navigate[0].Locator == "" {
		return "", fmt.Errorf("no URL defined in first locator for restaurant %s", r.Name)
	}

	url, err := url.Parse(r.Navigate[0].Locator)
	if err != nil {
		return "", fmt.Errorf("invalid URL in first locator for restaurant %s: %w", r.Name, err)
	}

	downloadPath, err = download.Curl(downloadPath, url.String())
	if err != nil {
		return "", fmt.Errorf("could not download file %s for restaurant %s: %w", url, r.Name, err)
	}

	logger.Info("Successfully downloaded menu", "path", downloadPath)
	return downloadPath, nil
}

func (r *Restaurant) Scrape(downloadPath string, web *web.Web, logger *slog.Logger) (string, error) {
	logger = r.withLogger(logger, "website", r.Website)
	logger.Info("Scraping restaurant")

	err := web.Run(r.Website, func(page playwright.Page) error {
		for i, nav := range r.Navigate {
			if nav.Style != "" {
				page.AddStyleTag(playwright.PageAddStyleTagOptions{Content: playwright.String(nav.Style)})
			}
			nav.Locator = placeholder.Replace(nav.Locator)
			selector := page.Locator(nav.Locator).First()
			if i < len(r.Navigate)-1 {
				logger.Debug("Clicking on locator", "locator", nav.Locator)
				time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
				if err := selector.Click(); err != nil {
					return fmt.Errorf("could not click on %s: %w", nav.Locator, err)
				}
			} else if r.ContentType != "html" {
				if nav.Attribute == "" {
					logger.Debug("Trying to download file by clicking on locator", "locator", nav.Locator)
					download, err := page.ExpectDownload(func() error {
						return selector.Click(playwright.LocatorClickOptions{Force: playwright.Bool(true)})
					})
					if err != nil {
						return fmt.Errorf("could not click on %s: %w", nav.Locator, err)
					}
					downloadPath = filepath.Join(DownloadsFolder, download.SuggestedFilename())
					if err := download.SaveAs(downloadPath); err != nil {
						return fmt.Errorf("could not save download to %s: %w", downloadPath, err)
					}
				} else {
					logger.Debug("Trying to download file by getting attribute", "locator", nav.Locator, "attribute", nav.Attribute)
					imgSrc, err := selector.GetAttribute(nav.Attribute)
					if err != nil {
						return fmt.Errorf("could not get attribute %s for restaurant %s: %w", nav.Attribute, r.Name, err)
					}
					downloadPath, err = download.Curl(downloadPath, imgSrc)
					if err != nil {
						return fmt.Errorf("could not download file %s for restaurant %s: %w", imgSrc, r.Name, err)
					}
				}
			} else {
				var err error
				time.Sleep(2 * time.Second)
				if nav.Locator != "" {
					logger.Debug("Making a screenshot of locator", "locator", nav.Locator)
					locator := page.Locator(nav.Locator).First()
					locatorCount, err := locator.Count()
					if err != nil {
						return fmt.Errorf("could not count locators %s: %w", nav.Locator, err)
					}
					if locatorCount == 0 {
						return fmt.Errorf("no element found for locator %s", nav.Locator)
					}
					err = locator.ScrollIntoViewIfNeeded()
					if err != nil {
						return fmt.Errorf("could not scroll: %w", err)
					}
					_, err = locator.Screenshot(playwright.LocatorScreenshotOptions{
						Animations: playwright.ScreenshotAnimationsDisabled,
						Path:       playwright.String(downloadPath),
						Type:       playwright.ScreenshotTypePng,
					})
				} else {
					logger.Debug("Making a screenshot of the full page")
					_, err = page.Screenshot(playwright.PageScreenshotOptions{
						Animations: playwright.ScreenshotAnimationsDisabled,
						Path:       playwright.String(downloadPath),
						FullPage:   playwright.Bool(true),
						Type:       playwright.ScreenshotTypePng,
					})
				}
				if err != nil {
					return fmt.Errorf("could not screenshot: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error scraping restaurant %s: %w", r.Name, err)
	}

	logger.Info("Successfully scraped restaurant", "path", downloadPath)
	return downloadPath, nil
}
