package services

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/playwright-community/playwright-go"
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

func (s *PlaywrightService) doScrape(url string, parse *Parse) (string, error) {
	slog.Debug("scraping url", "url", url)
	page, err := s.browser.NewPage(playwright.BrowserNewPageOptions{
		BypassCSP:         playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
		ReducedMotion:     playwright.ReducedMotionReduce,
	})
	if err != nil {
		return "", fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()
	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return "", fmt.Errorf("could not navigate to first page: %v", err)
	}
	downloadPath := fmt.Sprintf("%s%d", TempDownloadFolder, time.Now().Unix())
	for i, n := range parse.Navigate {
		n.Search = placeholder.Replace(n.Search)
		selector := page.Locator(n.Search).First()
		if i < len(parse.Navigate)-1 {
			slog.Debug("navigate", "search", n.Search)
			if err := selector.Click(); err != nil {
				return "", err
			}
			if err := selector.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}); err != nil {
				return "", err
			}
		} else if parse.IsFile {
			if n.Attribute == "" {
				slog.Debug("download", "search", n.Search)
				download, err := page.ExpectDownload(func() error {
					return selector.Click()
				})
				if err != nil {
					return "", err
				}
				downloadPath = TempDownloadFolder + download.SuggestedFilename()
				download.SaveAs(downloadPath)
			} else {
				slog.Debug("download", "attribute", n.Attribute)
				imgSrc, err := selector.GetAttribute(n.Attribute)
				if err != nil {
					return "", err
				}
				downloadPath, err = download.File(downloadPath, imgSrc)
				if err != nil {
					return "", err
				}
			}
		} else {
			slog.Debug("scroll into view", "search", n.Search)
			if err := selector.ScrollIntoViewIfNeeded(); err != nil {
				return "", err
			}
			time.Sleep(2 * time.Second)
		}
	}
	if !parse.IsFile {
		slog.Debug("screenshot", "downloadPath", downloadPath)
		page.Screenshot(playwright.PageScreenshotOptions{
			Animations: playwright.ScreenshotAnimationsDisabled,
			Path:       playwright.String(downloadPath),
			FullPage:   playwright.Bool(true),
			Type:       playwright.ScreenshotTypePng,
			Clip:       &playwright.Rect{X: parse.Clip.OffsetX, Y: parse.Clip.OffsetY, Width: parse.Clip.Width, Height: parse.Clip.Height},
		})
	}
	return downloadPath, nil
}
