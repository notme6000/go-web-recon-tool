package scraper

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Browser struct {
	browser *rod.Browser
}

func NewBrowser() (*Browser, error) {
	browser := rod.New().MustConnect()
	return &Browser{browser: browser}, nil
}

func (b *Browser) Close() error {
	b.browser.MustClose()
	return nil
}

func (b *Browser) ScrapeWebsite(url string) (string, error) {
	page := b.browser.MustPage(url)
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36",
	})
	page.MustWaitLoad()

	body := page.MustElement("body").MustText()
	return body, nil
}
