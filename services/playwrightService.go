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
	return &PlaywrightService{pw: pw}, nil
}

type PlaywrightService struct {
	browser            playwright.Browser
	pw                 *playwright.Playwright
	browserInitialized bool
}

func (s *PlaywrightService) close() {
	if s.browserInitialized {
		s.browser.Close()
	}
	s.pw.Stop()
}

func (s *PlaywrightService) initBrowser() error {
	if s.browserInitialized {
		return nil
	}
	browser, err := s.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("could not launch chromium: %w", err)
	}
	s.browser = browser
	s.browserInitialized = true
	return nil
}

func (s *PlaywrightService) doScrape(url string, parse *config.Parse) (string, error) {
	if err := s.initBrowser(); err != nil {
		return "", err
	}

	slog.Debug("scraping url", "url", url)
	page, err := s.browser.NewPage(playwright.BrowserNewPageOptions{
		BypassCSP:         playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	response, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return "", fmt.Errorf("could not navigate to first page: %v", err)
	}
	if response.Status() >= 400 {
		return "", fmt.Errorf(response.StatusText())
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
				locatorCount, err := locator.Count()
				if err != nil {
					return "", fmt.Errorf("could not count locators %s: %w", n.Locator, err)
				}
				if locatorCount == 0 {
					slog.Warn("locator not found on page", "locator", n.Locator)
				}
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
