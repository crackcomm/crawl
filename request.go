package crawl

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Request - HTTP Request.
// Multipart form is not implemented.
type Request struct {
	// URL - It can be absolute URL or a relative to source URL if referer is set.
	URL string `json:"url,omitempty"`
	// Method - "GET" by default.
	Method string `json:"method,omitempty"`
	// Referer - Request referer.
	Referer string `json:"referer,omitempty"`
	// Form - Form values which set as request body.
	Form url.Values `json:"form,omitempty"`
	// Query - Form values which set as url query.
	Query url.Values `json:"query,omitempty"`
	// Header - Header values.
	Header map[string]string `json:"header,omitempty"`
	// Raw - when set to false, it means we expect HTML response
	Raw bool `json:"raw,omitempty"`
	// Callbacks - Crawl callback list.
	Callbacks []string `json:"callbacks,omitempty"`
}

// Callbacks - Helper for creating list of strings (callback names).
func Callbacks(v ...string) []string {
	return v
}

// ConstructHTTPRequest - Constructs a http.Request structure.
func ConstructHTTPRequest(req *Request) (r *http.Request, err error) {
	u, err := req.ParseURL()
	if err != nil {
		return
	}

	// Make http request
	r = &http.Request{
		URL:    u,
		Method: req.GetMethod(),
		Header: make(http.Header),
	}

	if req.Form != nil {
		setRequestForm(req, r)
	}

	if req.Query != nil {
		r.URL.RawQuery = req.Query.Encode()
	}

	if r.Method == "" && req.Form != nil {
		r.Method = "POST"
	}

	for key, value := range req.Header {
		r.Header.Set(key, value)
	}

	// Set referer if any
	if req.Referer != "" {
		r.Header.Set("Referer", req.Referer)
	}

	return
}

func setRequestForm(req *Request, r *http.Request) {
	// Set Content-Type header
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set request method if currently is empty
	if req.Method == "" {
		r.Method = "POST"
		req.Method = "POST"
	}

	// Set request body
	data := strings.NewReader(req.Form.Encode())
	r.Body = ioutil.NopCloser(data)

	// Request content length
	r.ContentLength = int64(data.Len())
}

// ParseURL - Parses request URL.
// If request Source is set, parsed - URL is resolved
// with reference to source request URL.
func (req *Request) ParseURL() (u *url.URL, err error) {
	u, err = url.Parse(req.URL)
	if err != nil {
		return
	}
	if req.Referer != "" {
		ref, err := url.Parse(req.Referer)
		if err != nil {
			return nil, err
		}
		u = ref.ResolveReference(u)
	}
	return
}

// GetMethod - Returns request Method. If empty returns "GET".
func (req *Request) GetMethod() string {
	if req.Method == "" {
		return "GET"
	}
	return req.Method
}

// String - Returns "{method} {url}" formatted string.
func (req *Request) String() string {
	return fmt.Sprintf("%s %s", req.GetMethod(), req.URL)
}
