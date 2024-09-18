package services

import (
	"log/slog"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/ysmood/gson"
)

type Scraper struct {
	browser *rod.Browser
}

func NewScraper() *Scraper {
	return &Scraper{
		browser: rod.New().MustConnect(),
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

func (cdp *Scraper) Screenshot(url string, filePath string, parse Parse) error {
	slog.Info("creating screenshot", "url", url, "filePath", filePath)
	page := cdp.browser.MustPage(url).MustSetViewport(parse.Scan.Chrome.Width, parse.Scan.Chrome.Width, 1000, false).MustWaitLoad()

	for _, n := range parse.Navigate {
		page.MustElement(n).MustClick().MustWaitLoad()
	}
	img, err := page.ScrollScreenshot(&rod.ScrollScreenshotOptions{
		Format:  proto.PageCaptureScreenshotFormatPng,
		Quality: gson.Int(100),
	})
	if err != nil {
		return err
	}
	err = utils.OutputFile(filePath, img)
	if err != nil {
		return err
	}
	return nil
}

func (cdp *Scraper) DownloadFile(url string, filePath string, parse Parse) error {
	slog.Info("downloading file", "url", url, "filePath", filePath)
	page := cdp.browser.MustPage(url).MustWaitStable()
	wait := cdp.browser.MustWaitDownload()
	for _, n := range parse.Navigate {
		page.MustElementX(n).MustClick().MustWaitStable()
	}
	err := utils.OutputFile(filePath, wait())
	if err != nil {
		return err
	}
	return nil
}
