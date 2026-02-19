package services

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/download"
	"github.com/flohoss/mittagskarte/internal/placeholder"
	"github.com/playwright-community/playwright-go"
)

func newPlaywrightService() (*PlaywrightService, error) {
	if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
		return nil, fmt.Errorf("could not install driver: %w", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("could not launch chromium: %w", err)
	}
	return &PlaywrightService{browser: browser, pw: pw}, nil
}

type PlaywrightService struct {
	browser playwright.Browser
	pw      *playwright.Playwright
}

func (s *PlaywrightService) close() {
	if s.browser != nil {
		s.browser.Close()
	}
	s.pw.Stop()
}

func (s *PlaywrightService) doScrape(url string, parse *config.Parse) (string, error) {
	slog.Debug("scraping url", "url", url)
	page, err := s.browser.NewPage(playwright.BrowserNewPageOptions{
		BypassCSP:         playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	// Add anti-detection measures
	page.SetExtraHTTPHeaders(map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Accept-Language": "de-DE,de;q=0.9",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Encoding": "gzip, deflate, br",
		"DNT":             "1",
	})

	if err := page.AddInitScript(playwright.Script{
		Content: playwright.String("Object.defineProperty(navigator, 'webdriver', { get: () => false });"),
	}); err != nil {
		slog.Warn("could not add init script", "err", err)
	}

	// Add human-like delay
	time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)

	response, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return "", fmt.Errorf("could not navigate to first page: %v", err)
	}
	if response.Status() >= 400 {
		return "", fmt.Errorf("received non-success status code: %d", response.Status())
	}

	downloadPath := fmt.Sprintf("%s%d", TempDownloadFolder, time.Now().Unix())
	for i, n := range parse.Navigate {
		if n.Style != "" {
			page.AddStyleTag(playwright.PageAddStyleTagOptions{Content: playwright.String(n.Style)})
		}
		n.Locator = placeholder.Replace(n.Locator)
		selector := page.Locator(n.Locator).First()
		if i < len(parse.Navigate)-1 {
			slog.Debug("navigate", "locator", n.Locator)
			time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
			if err := selector.Click(); err != nil {
				return "", fmt.Errorf("could not click on %s: %w", n.Locator, err)
			}
		} else if parse.FileType == config.PDF || parse.FileType == config.Image {
			if n.Attribute == "" {
				slog.Debug("download", "locator", n.Locator)
				download, err := page.ExpectDownload(func() error {
					return selector.Click(playwright.LocatorClickOptions{Force: playwright.Bool(true)})
				})
				if err != nil {
					return "", fmt.Errorf("could not click on %s: %w", n.Locator, err)
				}
				downloadPath = TempDownloadFolder + download.SuggestedFilename()
				download.SaveAs(downloadPath)
			} else {
				slog.Debug("download", "locator", n.Locator, "attribute", n.Attribute)
				imgSrc, err := selector.GetAttribute(n.Attribute)
				if err != nil {
					return "", fmt.Errorf("could not get attribute %s: %w", n.Attribute, err)
				}
				downloadPath, err = download.Curl(downloadPath, imgSrc)
				if err != nil {
					return "", fmt.Errorf("could not download file %s: %w", imgSrc, err)
				}
			}
		} else {
			slog.Debug("screenshot", "downloadPath", downloadPath)
			var err error
			time.Sleep(2 * time.Second)
			if n.Locator != "" {
				slog.Debug("with locator", "locator", n.Locator)
				locator := page.Locator(n.Locator).First()
				err = locator.ScrollIntoViewIfNeeded()
				if err != nil {
					return "", fmt.Errorf("could not scroll: %w", err)
				}
				_, err = locator.Screenshot(playwright.LocatorScreenshotOptions{
					Animations: playwright.ScreenshotAnimationsDisabled,
					Path:       playwright.String(downloadPath),
					Type:       playwright.ScreenshotTypePng,
				})
			} else {
				slog.Debug("with full page")
				_, err = page.Screenshot(playwright.PageScreenshotOptions{
					Animations: playwright.ScreenshotAnimationsDisabled,
					Path:       playwright.String(downloadPath),
					FullPage:   playwright.Bool(true),
					Type:       playwright.ScreenshotTypePng,
				})
			}
			if err != nil {
				return "", fmt.Errorf("could not screenshot: %w", err)
			}
		}
	}
	return downloadPath, nil
}
