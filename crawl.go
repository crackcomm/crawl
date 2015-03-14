package crawl

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"golang.org/x/net/context"

	log "github.com/golang/glog"
)

// Options - Crawl options.
type Options struct {
	MaxRequestsPerMinute int // one request is send every MaxRequestsPerMinute / 60 seconds
	QueueCapacity        int // capacity of queue channel
}

// Crawl - Crawl structure.
// It keeps track of one-client crawl.
// Requests are scheduled in queue (look QueueCapacity option).
// Then depending on MaxRequestsPerMinute options they are executed.
// Each request should contain at least one Callback.
// Which is a interface{} key to Handler.
//
// HTTP Client proxy can be set in environment variable:
//
// HTTP_PROXY - for http requests,
// HTTPS_PROXY - for https requests,
// NO_PROXY - for no proxy at all
//
// Crawl can be Freeze()-ed when it's required.
type Crawl struct {
	*Options
	*Queue
	*http.Client

	mutex    *sync.RWMutex
	handlers map[interface{}]Handler

	closeCh chan bool // close channel
	doneCh  chan bool // done channel

	headers map[string]string // Default HTTP headers
}

// Handler - Crawl handler.
// Context contains crawl pointer which can be
// retrieved using crawl.FromContext method
type Handler func(context.Context, *Crawl, *Response) error

// DefaultOptions - Crawl default options.
// They are used when a new crawl is created.
var DefaultOptions = &Options{
	MaxRequestsPerMinute: 1000,
	QueueCapacity:        100000,
}

// New - Creates new crawl.
// Options are set permamently on Queue (QueueCapacity).
// When it's changed Queue should be remade.
// By default crawl is created with DefaultOptions
// If you want to change options set crawl.SetOptions() method.
func New() (crawl *Crawl) {
	c := &http.Client{
		Transport: http.DefaultTransport,
	}
	c.Jar, _ = cookiejar.New(nil)

	crawl = &Crawl{
		Client:   c,
		mutex:    new(sync.RWMutex),
		handlers: make(map[interface{}]Handler),
		closeCh:  make(chan bool, 1),
		doneCh:   make(chan bool, 1),
	}
	crawl.SetOptions(DefaultOptions)
	return
}

// SetOptions - Sets crawl options.
// If QueueCapacity option is changed, it creates a new Queue.
func (crawl *Crawl) SetOptions(opts *Options) {
	if opts == nil {
		return
	}

	// Compare queue capacity settings
	// Create new queue if settings are changed
	// Or current settings are empty
	if copts := crawl.Options; copts == nil ||
		(opts.QueueCapacity != 0 && copts.QueueCapacity != opts.QueueCapacity) {

		// Create queue with new capacity
		crawl.Queue = NewQueue(opts.QueueCapacity)
	}

	// Set options
	crawl.Options = opts
}

// Do - Makes http.Request using crawl.Client
// returns http.Response wrapped in Response structure.
func (crawl *Crawl) Do(req *Request) (resp *Response, err error) {
	// Get http.Request structure
	rq, err := req.HTTPRequest()
	if err != nil {
		return
	}

	// Copy default headers
	for k, v := range crawl.headers {
		if _, has := rq.Header[k]; !has {
			rq.Header.Set(k, v)
		}
	}

	// Make request
	r, err := crawl.Client.Do(rq)
	if err != nil {
		return
	}

	resp = &Response{Response: r}

	return
}

// Execute - Makes a http request using crawl.Client.
// If request HTML is set to true ParseHTML() method is executed on Response.
// Then all callbacks are executed with context containing crawl and response.
func (crawl *Crawl) Execute(req *Request) (resp *Response, err error) {
	// Make request
	resp, err = crawl.Do(req)
	if err != nil {
		return
	}

	// If request.Raw is not true - parse html
	if !req.Raw {
		err = resp.ParseHTML()
		if err != nil {
			return
		}
	}

	// Set request context if empty
	if req.Context == nil {
		req.Context = context.Background()
	}

	// ctx = context.WithValue(ctx, "crawl", crawl)
	// ctx = context.WithValue(ctx, "response", resp)

	// Run handlers
	for _, cb := range req.Callbacks {
		if handler := crawl.GetHandler(cb); handler != nil {
			err = handler(req.Context, crawl, resp)
			if err != nil {
				return
			}
		} else {
			log.Warningf("Handler %v was not found", cb)
		}
	}

	log.V(2).Infof("%s %s %s - %v", req.GetMethod(), resp.GetStatus(), resp.GetURL(), req.Callbacks)

	return
}

// Start - Starts reading from queue.
// After crawl will be done
func (crawl *Crawl) Start() (err error) {
	defer func() {
		crawl.doneCh <- true
	}()

	var tick <-chan time.Time
	if crawl.Options.MaxRequestsPerMinute > 0 {
		tick = time.Tick(time.Minute / time.Duration(crawl.Options.MaxRequestsPerMinute))
	}

	for crawl.Queue.Continue() {
		if tick != nil {
			<-tick
		}

		// Return if closed
		select {
		case <-crawl.closeCh:
			return
		default:
		}

		// Get and start request
		if request, ok := crawl.Queue.Get(); ok {
			go crawl.executeRequest(request)
		} else {
			return
		}
	}

	return
}

func (crawl *Crawl) executeRequest(req *Request) {
	if _, err := crawl.Execute(req); err != nil {
		log.Warningf("ERR: %v on %s", err, req.String())
	}
	crawl.Queue.Done()
}

// Schedule - Schedules request.
func (crawl *Crawl) Schedule(req *Request) {
	crawl.Queue.Schedule(req)
}

// Handler - Adds crawl handler.
// Handler is a callback referenced by name.
func (crawl *Crawl) Handler(name interface{}, h Handler) {
	crawl.mutex.Lock()
	crawl.handlers[name] = h
	crawl.mutex.Unlock()
}

// GetHandler - Gets crawl handler by name.
func (crawl *Crawl) GetHandler(name interface{}) Handler {
	crawl.mutex.RLock()
	defer crawl.mutex.RUnlock()
	return crawl.handlers[name]
}

// SetDefaultHeaders - Sets crawl default headers.
func (crawl *Crawl) SetDefaultHeaders(headers map[string]string) {
	crawl.headers = headers
}

// Done - Returns a channel that will be notified whenever
// crawl will be done, if it wasn't started - it will be never notified.
// It can be used when crawl is started in goroutine.
func (crawl *Crawl) Done() (done <-chan bool) {
	return crawl.doneCh
}

// Close - Closes crawl.
func (crawl *Crawl) Close() (err error) {
	crawl.Queue.Close()
	crawl.closeCh <- true
	return
}
