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
	URL       string                 `json:"url,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Referer   string                 `json:"referer,omitempty"`
	Form      map[string]string      `json:"form,omitempty"`
	Header    map[string]string      `json:"header,omitempty"`
	Raw       bool                   `json:"raw,omitempty"`
	Callbacks []string               `json:"callbacks,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
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

	for key, value := range req.Header {
		r.Header.Set(key, value)
	}

	// Set referer if any
	if req.Referer != "" {
		r.Header.Set("Referer", req.Referer)
	}

	// Return now if no form
	if req.Form != nil {
		setRequestForm(req, r)
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

// reqJob - Structure to make Request+Context a Job interface.
type reqJob struct {
	req *Request
	ctx context.Context
}

func (job *reqJob) Context() context.Context {
	return job.ctx
}

func (job *reqJob) Request() *Request {
	return job.req
}

func (job *reqJob) Done() {
}
