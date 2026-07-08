package web

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/mxschmitt/playwright-go"
)

func New() (*Web, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	return &Web{pw: pw}, nil
}

type Web struct {
	browser playwright.Browser
	pw      *playwright.Playwright

	mu                 sync.Mutex
	browserInitialized bool
}

func (s *Web) Close() {
	s.mu.Lock()
	browser := s.browser
	s.browser = nil
	s.browserInitialized = false
	s.mu.Unlock()

	if browser != nil {
		browser.Close()
	}
	s.pw.Stop()
}

func (s *Web) init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.browserInitialized && s.browser != nil && s.browser.IsConnected() {
		return nil
	}

	if s.browserInitialized && s.browser != nil {
		s.browser.Close()
		s.browserInitialized = false
	}

	browser, err := s.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("could not launch chromium: %w", err)
	}

	browser.OnDisconnected(func(_ playwright.Browser) {
		s.mu.Lock()
		s.browserInitialized = false
		s.browser = nil
		s.mu.Unlock()
		slog.Warn("Chromium browser disconnected; will relaunch on next scrape")
	})

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
	if response == nil {
		return fmt.Errorf("received nil response for %s", url)
	}
	if response.Status() >= 400 {
		return fmt.Errorf("received %d: %s", response.Status(), response.StatusText())
	}
	return fn(page)
}
