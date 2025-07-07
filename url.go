package main

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// clickLink performs an HTTP GET to the given link,
// mimicking a browser click by sending common headers,
// handling cookies, and following redirects.
// It returns the final response body as a string.
func clickLink(rawURL string) error {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("disable-dev-shm-usage", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()
	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(rawURL),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("html", &html),
	)
	return err
}