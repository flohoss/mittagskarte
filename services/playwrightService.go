package services

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/playwright-community/playwright-go"
	"gitlab.unjx.de/flohoss/mittag/config"
	"gitlab.unjx.de/flohoss/mittag/internal/download"
	"gitlab.unjx.de/flohoss/mittag/internal/placeholder"
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
	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	}); err != nil {
		return "", fmt.Errorf("could not navigate to first page: %v", err)
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
				downloadPath, err = download.File(downloadPath, imgSrc)
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
