package crawl

import (
	"net/url"
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

// Finder - HTML finder.
type Finder interface {
	Find(string) *goquery.Selection
}

// FindAny - Finds node in response and returns attr content.
func FindAny(finder Finder, selectors ...string) (node *goquery.Selection) {
	for _, selector := range selectors {
		node = finder.Find(selector)
		if node.Length() > 0 {
			break
		}
	}
	return
}

// Text - Finds node in response and returns text.
func Text(n Finder, selector string) string {
	return strings.Join(strings.Fields(getText(n.Find(selector))), " ")
}

func getText(n *goquery.Selection) string {
	return strings.Join(n.Map(func(_ int, s *goquery.Selection) string {
		return s.Text()
	}), " ")
}

// ParseFloat - Finds node in response and parses text as float64.
// When text is not found returns result 0.0 and nil error.
// Returned error source is strconv.ParseFloat.
func ParseFloat(n Finder, selector string) (res float64, err error) {
	if text := Text(n, selector); text != "" {
		text = strings.Replace(text, ",", ".", -1)
		res, err = strconv.ParseFloat(text, 64)
	}
	return
}

// ParseUint - Finds node in response and parses text as uint64.
// When text is not found returns result 0 and nil error.
// Returned error source is strconv.ParseUint.
func ParseUint(n Finder, selector string) (res uint64, err error) {
	if text := Text(n, selector); text != "" {
		text = strings.Replace(text, ",", "", -1)
		text = strings.Replace(text, " ", "", -1)
		res, err = strconv.ParseUint(text, 10, 64)
	}
	return
}

// NodeText - Returns node text.
// Helper for (*goquery.Selection).Each().
func NodeText(_ int, n *goquery.Selection) string {
	return strings.Join(strings.Fields(n.Text()), " ")
}

// Attr - Finds node in response and returns attr content.
func Attr(n Finder, attr, selector string) string {
	v, _ := n.Find(selector).Attr(attr)
	return strings.TrimSpace(v)
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
		return resp.URL().ResolveReference(u).String()
	}
}
