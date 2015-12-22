// Package utils implements utilities sometimes helpful for crawlers.
package utils

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/crackcomm/crawl"
)

var (
	// NodeHref -
	NodeHref = NodeAttr("href")

	// NodeDataPhoto -
	NodeDataPhoto = NodeAttr("data-photo")
)

// NodeText - Returns node text.
func NodeText(_ int, n *goquery.Selection) string {
	return n.Text()
}

// NodeAttr -  Returns node attribute selector.
func NodeAttr(attr string) func(int, *goquery.Selection) string {
	return func(_ int, n *goquery.Selection) (res string) {
		res, _ = n.Attr(attr)
		return
	}
}

// NodeResolveURL - Returns selector which takes href and resolves url.
func NodeResolveURL(resp *crawl.Response) func(int, *goquery.Selection) string {
	url := resp.GetURL()
	return func(_ int, n *goquery.Selection) (href string) {
		var ok bool
		href, ok = n.Attr("href")
		if !ok {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		return url.ResolveReference(u).String()
	}
}
