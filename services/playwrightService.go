package services

import (
	"fmt"
	"os"

	"github.com/playwright-community/playwright-go"
)

func newPlaywrightService(options SiteOptions) (*PlaywrightService, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch chromium: %w", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}
	if _, err = page.Goto(options.url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return nil, fmt.Errorf("could not navigate to first page: %v", err)
	}

	return &PlaywrightService{
		options:    options,
		browser:    browser,
		page:       page,
		Playwright: pw,
	}, nil
}

type PlaywrightService struct {
	options    SiteOptions
	browser    playwright.Browser
	page       playwright.Page
	Playwright *playwright.Playwright
}

type SiteOptions struct {
	url      string
	id       string
	parse    *Parse
	rawPath  string
	filePath string
}

func (s *PlaywrightService) close() {
	if s.browser != nil {
		s.browser.Close()
	}
	s.Playwright.Stop()
}

func (s *PlaywrightService) doScrape() error {
	for _, n := range s.options.parse.Navigate {
		if n.Attribute == "" {
			s.page.Locator(n.Search).Click()
		} else {
			download, err := s.page.ExpectDownload(func() error {
				return s.page.Locator(n.Search).Click()
			})
			if err != nil {
				return err
			}
			fileName := download.SuggestedFilename()
			download.SaveAs("tmp/downloads/" + fileName)
			os.Rename(fileName, s.options.rawPath)
		}
	}
	return nil
}
