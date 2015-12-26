package crawl

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NodeHref -
var NodeHref = NodeAttr("href")

// NodeDataPhoto -
var NodeDataPhoto = NodeAttr("data-photo")

// Text - Finds node in response and returns text.
func Text(resp *Response, selector string) string {
	return strings.TrimSpace(resp.Query().Find(selector).Text())
}

// NodeText - Returns node text.
// Helper for (*goquery.Selection).Each().
func NodeText(_ int, n *goquery.Selection) string {
	return n.Text()
}

// NodeAttr -  Returns node attribute selector.
// Helper for (*goquery.Selection).Each().
func NodeAttr(attr string) func(int, *goquery.Selection) string {
	return func(_ int, n *goquery.Selection) (res string) {
		res, _ = n.Attr(attr)
		return
	}
}

// NodeResolveURL - Returns selector which takes href and resolves url.
// Returns helper for (*goquery.Selection).Each().
func NodeResolveURL(resp *Response) func(int, *goquery.Selection) string {
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
