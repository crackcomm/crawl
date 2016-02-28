package crawl

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Response - Crawl http response.
// It is expected it to be a HTML response but not required.
// It ALWAYS has to be released using Close() method.
type Response struct {
	*Request
	*http.Response
	doc  *goquery.Document
	body *bytes.Buffer
}

var buffersPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// ParseHTML - Reads response body and parses it as HTML.
func (r *Response) ParseHTML() (err error) {
	body, err := r.BodyBuffer()
	if err != nil {
		return
	}
	r.doc, err = goquery.NewDocumentFromReader(body)
	return
}

// BodyBuffer - Reads response body and returns body buffer.
func (r *Response) BodyBuffer() (body *bytes.Buffer, err error) {
	if r.body == nil {
		r.body = buffersPool.Get().(*bytes.Buffer)
		r.body.Reset()
		_, err = io.Copy(r.body, r.Response.Body)
		if err != nil {
			return nil, err
		}
		// close response body
		r.Response.Body.Close()
	}
	return r.body, nil
}

// BodyBytes - Reads response body and returns byte array.
func (r *Response) BodyBytes() (body []byte, err error) {
	b, err := r.BodyBuffer()
	if err != nil {
		return
	}
	return b.Bytes(), nil
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
	if r.body != nil {
		buffersPool.Put(r.body)
	}
	// close response body
	// even it should be closed after a read
	// but to make sure we can just close again
	r.Response.Body.Close()
	return nil
}
