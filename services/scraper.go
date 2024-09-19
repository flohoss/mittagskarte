package services

import (
	"log/slog"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/cdp"
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

type Scraper struct {
	browser *rod.Browser
}

func NewScraper() *Scraper {
	cdp := cdp.New().
		Logger(utils.Log(func(args ...interface{}) {
			switch v := args[0].(type) {
			case *cdp.Request:
				slog.Debug("Request", "ID", v.ID, "method", v.Method, "url", v.Params)
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
	page := cdp.browser.MustPage(url).Timeout(2 * time.Minute).MustWaitStable()
	info := page.MustInfo()

	for _, n := range parse.Navigate {
		slog.Debug("navigating", "url", info.URL)
		selectTheRightMethod(page, n).Click(proto.InputMouseButtonLeft, 1)
		page.MustWaitStable()
	}

	slog.Debug("making screenshot", "url", info.URL)
	img, err := page.MustSetViewport(parse.Scan.ViewportWidth, 0, 1, false).ScrollScreenshot(&rod.ScrollScreenshotOptions{
		Format:      proto.PageCaptureScreenshotFormatPng,
		FixedTop:    parse.Scan.FixedTop,
		FixedBottom: parse.Scan.FixedBottom,
	})
	if err != nil {
		return err
	}

	slog.Debug("saving screenshot", "filePath", filePath)
	if err := utils.OutputFile(filePath, img); err != nil {
		return err
	}
	return nil
}

func (cdp *Scraper) DownloadFile(url string, filePath string, parse Parse) error {
	page := cdp.browser.MustPage(url).Timeout(30 * time.Second).MustWaitStable()

	for _, n := range parse.Navigate {
		slog.Debug("navigating", "url", page.MustInfo().URL)
		page.MustSearch(n.Search).MustClick().MustWaitStable()
	}

	wait := cdp.browser.MustWaitDownload()

	if err := utils.OutputFile(filePath, wait); err != nil {
		return err
	}
	return nil
}
