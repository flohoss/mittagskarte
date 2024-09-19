package services

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/cdp"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

type SearchBy string

const (
	CSS   SearchBy = "css"
	XPath SearchBy = "x-path"
	Name  SearchBy = "name"
	Regex SearchBy = "regex"
	JS    SearchBy = "js"
)

type Scraper struct {
	browser *rod.Browser
}

func NewScraper() *Scraper {
	cdp := cdp.New().
		Logger(utils.Log(func(args ...interface{}) {
			switch v := args[0].(type) {
			case *cdp.Request:
				slog.Debug(fmt.Sprintf("%3d", v.ID), "method", v.Method, "params", v.Params)
			}
		})).
		Start(cdp.MustConnectWS(launcher.New().MustLaunch()))

	return &Scraper{
		browser: rod.New().Client(cdp).MustConnect(),
	}
}

func (cdp *Scraper) Close() {
	if cdp.browser != nil {
		cdp.browser.MustClose()
	}
}

func (cdp *Scraper) NavigateToFinalUrl(parse Parse) {
	if len(parse.Navigate) == 0 {
		return
	}
}

func selectTheRightMethod(page *rod.Page, selector Selector) *rod.Element {
	switch selector.SearchBy {
	case CSS:
		slog.Debug("searching by css-selector", "search", selector.Search)
		return page.MustElement(selector.Search)
	case XPath:
		slog.Debug("searching by xpath-selector", "search", selector.Search)
		return page.MustElementX(selector.Search)
	case Name:
		slog.Debug("searching by name-selector", "search", selector.Search)
		return page.MustSearch(selector.Search)
	case Regex:
		slog.Debug("searching by regex-selector", "regex", selector.Regex)
		return page.MustElementR(selector.Search, selector.Regex)
	default:
		slog.Debug("no valid selector")
		os.Exit(1)
		return nil
	}
}

func (cdp *Scraper) Screenshot(url string, filePath string, parse Parse) error {
	opts := proto.TargetCreateTarget{
		URL: url,
	}
	page, err := cdp.browser.Page(opts)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}

	if err := page.Timeout(2 * time.Minute).WaitStable(5 * time.Second); err != nil {
		return fmt.Errorf("failed to wait for page stability: %w", err)
	}

	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("failed to get page info: %w", err)
	}

	for _, n := range parse.Navigate {
		slog.Debug("navigating", "url", info.URL)
		if err := selectTheRightMethod(page, n).Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("failed to click element: %w", err)
		}
		if err := page.WaitStable(5 * time.Second); err != nil {
			return fmt.Errorf("failed to wait for page stability after click: %w", err)
		}
	}

	slog.Debug("making screenshot", "url", info.URL)
	img, err := page.MustSetViewport(parse.Scan.ViewportWidth, 0, 1, false).ScrollScreenshot(&rod.ScrollScreenshotOptions{
		Format:      proto.PageCaptureScreenshotFormatPng,
		FixedTop:    parse.Scan.FixedTop,
		FixedBottom: parse.Scan.FixedBottom,
	})
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	slog.Debug("saving screenshot", "filePath", filePath)
	if err := utils.OutputFile(filePath, img); err != nil {
		return fmt.Errorf("failed to save screenshot to file: %w", err)
	}
	return nil
}

func (cdp *Scraper) DownloadFile(url string, filePath string, parse Parse) error {
	opts := proto.TargetCreateTarget{
		URL: url,
	}

	page, err := cdp.browser.Page(opts)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}

	if err := page.Timeout(2 * time.Minute).WaitStable(5 * time.Second); err != nil {
		return fmt.Errorf("failed to wait for page stability: %w", err)
	}

	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("failed to get page info: %w", err)
	}

	for _, n := range parse.Navigate {
		if n.Attribute != "" {
			slog.Debug("downloading file", "url", info.URL)
			attrValue, err := selectTheRightMethod(page, n).Attribute(n.Attribute)
			if err != nil {
				return fmt.Errorf("failed to get attribute %s: %w", n.Attribute, err)
			}

			fullURL := fmt.Sprintf("%s%s", n.Prefix, *attrValue)
			if err := fetch.DownloadFile(filePath, fullURL); err != nil {
				return fmt.Errorf("failed to download file from URL %s: %w", fullURL, err)
			}
		} else {
			slog.Debug("navigating", "url", info.URL)
			if err := selectTheRightMethod(page, n).Click(proto.InputMouseButtonLeft, 1); err != nil {
				return fmt.Errorf("failed to click element: %w", err)
			}
			if err := page.WaitStable(5 * time.Second); err != nil {
				return fmt.Errorf("failed to wait for page stability after click: %w", err)
			}
		}
	}
	return nil
}
