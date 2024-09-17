package services

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

type ChromeDp struct {
	timeoutCtx       *context.Context
	timeoutCtxCancel func()
	ctx              *context.Context
	cancel           func()
}

func NewChromeDp() *ChromeDp {
	cdp := &ChromeDp{}
	cdp.createContext()
	return cdp
}

func (cdp *ChromeDp) Close() {
	if cdp.ctx != nil {
		cdp.cancel()
	}
	if cdp.timeoutCtx != nil {
		cdp.timeoutCtxCancel()
	}
}

func (cdp *ChromeDp) createContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	cdp.timeoutCtx = &ctx
	cdp.timeoutCtxCancel = cancel

	ctx, cancel = chromedp.NewContext(ctx)
	cdp.ctx = &ctx
	cdp.cancel = cancel
}

func (cdp *ChromeDp) NavigateToFinalUrl(url string, finalUrl *string, parse Parse) error {
	if len(parse.Navigate) == 0 {
		*finalUrl = url
		return nil
	}

	var ok bool
	for _, click := range parse.Navigate {
		slog.Debug("searching url", "url", url)
		if err := chromedp.Run(*cdp.ctx, chromedp.Tasks{chromedp.Navigate(url), chromedp.AttributeValue(click.JQuery, click.Attribute, finalUrl, &ok)}); err != nil {
			return err
		}
		if click.Prefix != "" {
			*finalUrl = click.Prefix + *finalUrl
		}
	}
	return nil
}

func (cdp *ChromeDp) Screenshot(url, filePath string, chrome Chrome, parse Parse) error {
	var buf []byte
	slog.Debug("navigating url", "url", url)
	if err := chromedp.Run(*cdp.ctx, cdp.fullScreenshot(url, &buf, chrome.Width, parse.Hide)); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		return err
	}
	return nil
}

func (cdp *ChromeDp) fullScreenshot(urlstr string, res *[]byte, width int64, hideSelectors []string) chromedp.Tasks {
	tasks := chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(width, 0, 1.0, false),
		chromedp.Navigate(urlstr),
		chromedp.Sleep(2 * time.Second),
	}
	for _, hide := range hideSelectors {
		tasks = append(tasks, chromedp.Evaluate(hide, nil), chromedp.Sleep(1*time.Second))
	}
	return append(tasks, chromedp.FullScreenshot(res, 100))
}
