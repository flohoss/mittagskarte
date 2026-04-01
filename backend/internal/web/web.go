package web

import (
	"fmt"
	"log/slog"

	"github.com/playwright-community/playwright-go"
)

func New() (*Web, error) {
	if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
		return nil, fmt.Errorf("could not install driver: %w", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	return &Web{pw: pw}, nil
}

type Web struct {
	browser            playwright.Browser
	pw                 *playwright.Playwright
	browserInitialized bool
}

func (s *Web) Close() {
	if s.browserInitialized {
		s.browser.Close()
	}
	s.pw.Stop()
}

func (s *Web) init() error {
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

func (s *Web) Run(url string, fn func(page playwright.Page) error) error {
	if err := s.init(); err != nil {
		return err
	}

	slog.Debug("scraping url", "url", url)
	page, err := s.browser.NewPage(playwright.BrowserNewPageOptions{
		BypassCSP:         playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	response, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return fmt.Errorf("could not navigate to page: %v", err)
	}
	if response.Status() >= 400 {
		return fmt.Errorf("received %d: %s", response.Status(), response.StatusText())
	}
	return fn(page)
}
