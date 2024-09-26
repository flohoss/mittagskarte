package services

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
)

type SearchBy string

const (
	CSS   SearchBy = "css"
	XPath SearchBy = "x-path"
	Name  SearchBy = "name"
	Regex SearchBy = "regex"
	JS    SearchBy = "js"
)

type ScraperService struct {
	id       string
	panicked chan bool
	launcher *launcher.Launcher
	browser  *rod.Browser
	page     *rod.Page
}

func NewScraperService(id string) *ScraperService {
	slog.Debug("new scraper service", "id", id)
	launcher := launcher.New()
	u := launcher.Headless(true).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	panicked := make(chan bool)

	browser = browser.MustIncognito().MustIgnoreCertErrors(true).WithPanic(func(v interface{}) {
		slog.Error("browser panic", "id", id, "err", v)
		panicked <- true
		runtime.Goexit()
	})

	return &ScraperService{
		id:       id,
		panicked: panicked,
		launcher: launcher,
		browser:  browser,
	}
}

func (s *ScraperService) Close() {
	if s.browser != nil {
		s.browser.MustClose()
		slog.Debug("browser closed", "id", s.id)
	}
}

func (s *ScraperService) selectTheRightMethod(selector *Selector) (*rod.Element, error) {
	timeout := 10 * time.Second
	var el *rod.Element
	var err error
	switch selector.SearchBy {
	case CSS:
		slog.Debug("searching by css-selector", "id", s.id, "search", selector.Search)
		el, err = s.page.Timeout(timeout).Element(selector.Search)
		break
	case XPath:
		slog.Debug("searching by xpath-selector", "id", s.id, "search", selector.Search)
		el, err = s.page.Timeout(timeout).ElementX(selector.Search)
		break
	case Name:
		slog.Debug("searching by name-selector", "id", s.id, "search", selector.Search)
		var res *rod.SearchResult
		res, err = s.page.Timeout(timeout).Search(selector.Search)
		el = res.First
		break
	case Regex:
		slog.Debug("searching by regex-selector", "id", s.id, "regex", selector.Regex)
		el, err = s.page.Timeout(timeout).ElementR(selector.Search, selector.Regex)
		break
	default:
		slog.Debug("no valid selector", "id", s.id)
		os.Exit(1)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to select the element (id: %s): %w", s.id, err)
	}
	return el, nil
}

func (s *ScraperService) navigateToFirstPage(url string) {
	slog.Debug("navigating to first page", "id", s.id, "url", url)
	s.page = s.browser.Timeout(5 * time.Minute).MustPage(url).MustWaitStable()
}

func (s *ScraperService) navigateToAction(n *Selector) error {
	el, err := s.selectTheRightMethod(n)
	if err != nil {
		return err
	}
	slog.Debug("clicking on element", "id", s.id)
	el.MustWaitInteractable().MustClick().MustWaitInvisible()
	return nil
}

func (s *ScraperService) screenshot(url string, filePath string, parse Parse) error {
	for _, n := range parse.Navigate {
		if err := s.navigateToAction(&n); err != nil {
			return err
		}
	}
	slog.Debug("setting viewport", "id", s.id)
	s.page = s.page.MustSetViewport(parse.Scan.ViewportWidth, 0, 1, false)
	slog.Debug("taking screenshot", "id", s.id)
	img, err := s.page.ScrollScreenshot(&rod.ScrollScreenshotOptions{
		Format:      proto.PageCaptureScreenshotFormatPng,
		FixedTop:    parse.Scan.FixedTop,
		FixedBottom: parse.Scan.FixedBottom,
	})
	if err != nil {
		return fmt.Errorf("failed to take screenshot (id: %s): %w", s.id, err)
	}
	slog.Debug("saving screenshot", "id", s.id, "path", filePath)
	if err := utils.OutputFile(filePath, img); err != nil {
		return fmt.Errorf("failed to save screenshot (id: %s): %w", s.id, err)
	}
	return nil
}

func (s *ScraperService) downloadFile(url string, filePath string, parse Parse) error {
	for _, n := range parse.Navigate {
		if n.Attribute == "" {
			if err := s.navigateToAction(&n); err != nil {
				return err
			}
		} else {
			wait := s.browser.MustWaitDownload()
			el, err := s.selectTheRightMethod(&n)
			if err != nil {
				return err
			}
			slog.Debug("clicking on element", "id", s.id)
			el.MustClick()
			slog.Debug("downloading file", "id", s.id, "path", filePath)
			if err := utils.OutputFile(filePath, wait()); err != nil {
				return fmt.Errorf("failed to download file: %w", err)
			}
		}
	}
	return nil
}
