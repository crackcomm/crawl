package crawl

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
)

// Request - HTTP Request.
// Method is "GET" by default.
// URL can be absolute URL or a relative to source URL.
// If Raw is set to false, it means we expect HTML response
// and crawler, if request will be successfull will parse
// Response body as HTML.
// If Form is not empty, request will be send with POST method (if empty)
// and with form encoded body.
// Multipart form is not implemented.
type Request struct {
	Method, URL string
	Raw         bool
	Source      *Response
	Callbacks   []interface{}
	Form        map[string]string
	Context     context.Context
}

// Callbacks - Helper for creating list of interfaces.
func Callbacks(v ...interface{}) []interface{} {
	return v
}

// HTTPRequest - Creates a http.Request structure.
// If Method is empty it's set "GET".
func (req *Request) HTTPRequest() (r *http.Request, err error) {
	u, err := req.GetURL()
	if err != nil {
		return
	}

	// Make http request
	r = &http.Request{
		URL:    u,
		Method: req.GetMethod(),
		Header: make(http.Header),
	}

	// Set referer if any
	if req.Source != nil {
		r.Header.Set("Referer", req.Source.GetURL().String())
	}

	// Return now if no form
	if req.Form == nil {
		return
	}

	// Set Content-Type header
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set request method if currently is empty
	if req.Method == "" {
		r.Method = "POST"
		req.Method = "POST"
	}

	// Copy form values to request form
	formdata := make(url.Values)
	for key, value := range req.Form {
		formdata.Set(key, value)
	}

	// Set request body
	data := strings.NewReader(formdata.Encode())
	r.Body = ioutil.NopCloser(data)

	// Request content length
	r.ContentLength = int64(data.Len())

	return
}

// GetMethod - Returns request Method. If empty returns "GET".
func (req *Request) GetMethod() string {
	if req.Method == "" {
		return "GET"
	}
	return req.Method
}

// ResolveURL - Resolves url
func (req *Request) ResolveURL(uri *url.URL) (u *url.URL, err error) {
	u, err = req.GetURL()
	if uri != nil {
		u = u.ResolveReference(uri)
	}
	return
}

// GetURL - Parses request URL.
// If request Source is set, parsed - URL is resolved
// with reference to source request URL.
func (req *Request) GetURL() (u *url.URL, err error) {
	u, err = url.Parse(req.URL)
	if err != nil {
		return
	}
	if src := req.Source; src != nil {
		u = src.GetURL().ResolveReference(u)
	}
	return
}

// String -
func (req *Request) String() string {
	return fmt.Sprintf("%s %s", req.GetMethod(), req.URL)
}
