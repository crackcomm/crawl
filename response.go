package crawl

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// Response - Crawl http response.
// It is expected it to be a HTML response but not required.
// It ALWAYS has to be released using Close() method.
type Response struct {
	*Request
	*http.Response
	doc  *goquery.Document
	body []byte
}

// ParseHTML - Reads response body and parses it as HTML.
func (r *Response) ParseHTML() (err error) {
	body, err := r.Bytes()
	if err != nil {
		return
	}
	r.doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(body))
	return
}

// Bytes - Reads response body and returns byte array.
func (r *Response) Bytes() (body []byte, err error) {
	if r.body == nil {
		err = r.readBody()
	}
	return r.body, err
}

// Status - Gets response status.
func (r *Response) Status() string {
	return r.Response.Status
}

// URL - Gets response request URL.
func (r *Response) URL() *url.URL {
	return r.Response.Request.URL
}

// Query - Returns goquery.Document.
func (r *Response) Query() *goquery.Document {
	return r.doc
}

// Find - Short for: r.Query().Find(selector).
func (r *Response) Find(selector string) *goquery.Selection {
	return r.doc.Find(selector)
}

// Close - Closes response body.
func (r *Response) Close() error {
	// close response body
	// even though it should be closed after a read
	// but to make sure we can just close again
	return r.Response.Body.Close()
}

// readBody - Reads response body to `r.body`.
func (r *Response) readBody() (err error) {
	if r.body != nil {
		return
	}
	defer r.Response.Body.Close()
	r.body, err = ioutil.ReadAll(r.Response.Body)
	return
}
