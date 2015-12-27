package crawl

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NodeHref - Node "href" attribute selector.
var NodeHref = NodeAttr("href")

// NodeSrc - Node "src" attribute selector.
var NodeSrc = NodeAttr("src")

// NodeDataPhoto - Node "data-photo" attribute selector.
var NodeDataPhoto = NodeAttr("data-photo")

// FindAny - Finds node in response and returns attr content.
func FindAny(resp *Response, selectors ...string) (node *goquery.Selection) {
	for _, selector := range selectors {
		node = resp.Query().Find(selector)
		if node.Length() > 0 {
			break
		}
	}
	return
}

// Text - Finds node in response and returns text.
func Text(resp *Response, selector string) string {
	return strings.TrimSpace(resp.Query().Find(selector).Text())
}

// ParseFloat - Finds node in response and parses text as float64.
// When text is not found returns result 0.0 and nil error.
// Returned error source is strconv.ParseFloat.
func ParseFloat(resp *Response, selector string) (res float64, err error) {
	if text := Text(resp, selector); text != "" {
		text = strings.Replace(text, ",", ".", -1)
		res, err = strconv.ParseFloat(text, 64)
	}
	return
}

// Attr - Finds node in response and returns attr content.
func Attr(resp *Response, attr, selector string) string {
	v, _ := resp.Query().Find(selector).Attr(attr)
	return strings.TrimSpace(v)
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
