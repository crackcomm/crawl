package crawl

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Response - Crawl http response.
// It is expected it to be a HTML response but not required.
// It ALWAYS has to be released using Close() method.
type Response struct {
	*http.Response
	Body []byte

	doc *goquery.Document
}

// Text - Finds node in response and returns text.
func Text(resp *Response, selector string) string {
	return strings.TrimSpace(resp.Query().Find(selector).Text())
}

// ParseHTML - Reads response body and parses it as HTML.
func (r *Response) ParseHTML() (err error) {
	err = r.Read()
	if err != nil {
		return
	}

	rdr := bytes.NewBuffer(r.Body)
	r.doc, err = goquery.NewDocumentFromReader(rdr)
	return
}

// Read - Reads response body into Body.
// If body already was readen it won't do nothing.
func (r *Response) Read() (err error) {
	if r.Body == nil {
		r.Body, err = ioutil.ReadAll(r.Response.Body)
	}
	return
}

// WriteFile - Reads response body to memory and writes to file.
func (r *Response) WriteFile(fname string) (err error) {
	// Read response body
	err = r.Read()
	if err != nil {
		return
	}
	err = ioutil.WriteFile(fname, r.Body, os.ModePerm)
	return

}

// GetURL - Gets response request URL.
func (r *Response) GetURL() *url.URL {
	return r.GetRequest().URL
}

// GetRequest - Gets response http.Request source.
func (r *Response) GetRequest() *http.Request {
	return r.Response.Request
}

// GetStatus - Gets response status.
func (r *Response) GetStatus() string {
	return r.Response.Status
}

// Query - Returns goquery.Document.
func (r *Response) Query() *goquery.Document {
	return r.doc
}

// Close - Closes response body.
func (r *Response) Close() (err error) {
	return r.Response.Body.Close()
}
