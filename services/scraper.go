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
	page    *rod.Page
	info    *proto.TargetTargetInfo
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

func (cdp *Scraper) selectTheRightMethod(selector *Selector) (*rod.Element, error) {
	switch selector.SearchBy {
	case CSS:
		slog.Debug("searching by css-selector", "search", selector.Search)
		return cdp.page.Element(selector.Search)
	case XPath:
		slog.Debug("searching by xpath-selector", "search", selector.Search)
		return cdp.page.ElementX(selector.Search)
	case Name:
		slog.Debug("searching by name-selector", "search", selector.Search)
		res, err := cdp.page.Search(selector.Search)
		if err != nil {
			return nil, err
		}
		return res.First, nil
	case Regex:
		slog.Debug("searching by regex-selector", "regex", selector.Regex)
		return cdp.page.ElementR(selector.Search, selector.Regex)
	default:
		slog.Debug("no valid selector")
		os.Exit(1)
	}
	return nil, nil
}

func (cdp *Scraper) navigateToFirstPage(url string) error {
	opts := proto.TargetCreateTarget{
		URL: url,
	}
	page, err := cdp.browser.Page(opts)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}
	cdp.page = page.Timeout(1 * time.Minute)

	if err := cdp.page.WaitStable(5 * time.Second); err != nil {
		return fmt.Errorf("failed to wait for page stability: %w", err)
	}

	info, err := cdp.page.Info()
	if err != nil {
		return fmt.Errorf("failed to get page info: %w", err)
	}
	cdp.info = info
	return nil
}

func (cdp *Scraper) navigateToAction(n *Selector) error {
	slog.Debug("navigating", "url", cdp.info.URL)
	el, err := cdp.selectTheRightMethod(n)
	if err != nil {
		return fmt.Errorf("failed to get element: %w", err)
	}
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click element: %w", err)
	}
	if err := cdp.page.WaitStable(5 * time.Second); err != nil {
		return fmt.Errorf("failed to wait for page stability after click: %w", err)
	}
	return nil
}

func (cdp *Scraper) screenshot(url string, filePath string, parse Parse) error {
	for _, n := range parse.Navigate {
		if err := cdp.navigateToAction(&n); err != nil {
			return fmt.Errorf("failed to navigate to action: %w", err)
		}
	}

	if err := cdp.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Height:      0,
		Mobile:      false,
		ScreenWidth: &parse.Scan.ViewportWidth,
	}); err != nil {
		return fmt.Errorf("failed to set viewport: %w", err)
	}

	slog.Debug("making screenshot", "url", cdp.info.URL)
	img, err := cdp.page.ScrollScreenshot(&rod.ScrollScreenshotOptions{
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

func (cdp *Scraper) downloadFile(url string, filePath string, parse Parse) error {
	for _, n := range parse.Navigate {
		if n.Attribute == "" {
			if err := cdp.navigateToAction(&n); err != nil {
				return fmt.Errorf("failed to navigate to action: %w", err)
			}
		} else {
			slog.Debug("downloading file", "url", cdp.info.URL)
			el, err := cdp.selectTheRightMethod(&n)
			if err != nil {
				return fmt.Errorf("failed to get element: %w", err)
			}
			attrValue, err := el.Attribute(n.Attribute)
			if err != nil {
				return fmt.Errorf("failed to get attribute %s: %w", n.Attribute, err)
			}

			fullURL := fmt.Sprintf("%s%s", n.Prefix, *attrValue)
			if err := fetch.DownloadFile(filePath, fullURL); err != nil {
				return fmt.Errorf("failed to download file from URL %s: %w", fullURL, err)
			}
		}
	}
	return nil
}
