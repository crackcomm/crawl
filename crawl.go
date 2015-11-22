package crawl

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// Options - Crawl options.
type Options struct {
	MaxRequestsPerMinute int
	MaxRequestsPerSecond int

	QueueCapacity    int // capacity of queue channel
	QueueConcurrency int
}

// Crawl - Crawl structure.
// It keeps track of one-client crawl.
// Requests are scheduled in queue (look QueueCapacity option).
// Then depending on MaxRequestsPerSecond options they are executed.
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
	handlers map[interface{}][]Handler

	closeCh chan bool // close channel
	doneCh  chan bool // done channel

	headers map[string]string // Default HTTP headers
}

// Handler - Crawl handler.
type Handler func(context.Context, *Response) error

// DefaultOptions - Crawl default options.
// They are used when a new crawl is created.
var DefaultOptions = &Options{
	MaxRequestsPerSecond: 500,
	QueueCapacity:        100000,
	QueueConcurrency:     100,
}

// New - Creates new crawl.
// Options are set permamently on Queue (QueueCapacity).
// When it's changed Queue should be remade.
// By default crawl is created with DefaultOptions
// If you want to change options set crawl.SetOptions() method.
// Only one opts is accepted, rest is ignored.
func New(opts ...*Options) (crawl *Crawl) {
	c := &http.Client{
		Transport: http.DefaultTransport,
	}
	c.Jar, _ = cookiejar.New(nil)

	crawl = &Crawl{
		Client:   c,
		mutex:    new(sync.RWMutex),
		handlers: make(map[interface{}][]Handler),
		closeCh:  make(chan bool, 1),
		doneCh:   make(chan bool, 1),
	}
	if len(opts) == 0 {
		crawl.SetOptions(DefaultOptions)
	} else {
		crawl.SetOptions(opts[0])
	}
	return
}

// SetOptions - Sets crawl options.
// If QueueCapacity option is changed, it creates a new Queue.
func (crawl *Crawl) SetOptions(options *Options) {
	if options == nil {
		options = new(Options)
	}
	if options.MaxRequestsPerSecond == 0 {
		options.MaxRequestsPerSecond = DefaultOptions.MaxRequestsPerSecond
	}
	if options.QueueCapacity == 0 {
		options.QueueCapacity = DefaultOptions.QueueCapacity
	}
	if options.QueueConcurrency == 0 {
		options.QueueConcurrency = DefaultOptions.QueueConcurrency
	}

	// Compare queue capacity settings
	// Create new queue if settings are changed
	// Or current settings are empty
	if current := crawl.Options; current == nil || current.QueueCapacity < options.QueueCapacity {
		crawl.Queue = NewQueue(options.QueueCapacity)
	}

	// Set options
	crawl.Options = options
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

	// Send request and read response
	res, err := crawl.Client.Do(rq)
	if err != nil {
		return
	}

	return &Response{Response: res}, nil
}

// Execute - Makes a http request using crawl.Client.
// If request HTML is set to true ParseHTML() method is executed on Response.
// Then all callbacks are executed with context containing crawl and response.
func (crawl *Crawl) Execute(req *Request) (resp *Response, err error) {
	// Send request and read response
	resp, err = crawl.Do(req)
	if err != nil {
		return
	}

	// Parse HTML if not request.Raw
	if !req.Raw {
		err = resp.ParseHTML()
		if err != nil {
			return
		}
	}

	// Set new request context if empty
	if req.Context == nil {
		req.Context = context.Background()
	}

	// Run handlers
	for _, cb := range req.Callbacks {
		if handlers := crawl.GetHandlers(cb); len(handlers) >= 1 {
			for _, handler := range handlers {
				err = handler(req.Context, resp)
				if err != nil {
					return
				}
			}
		} else {
			return nil, fmt.Errorf("Handlers for %v was not found", cb)
		}
	}

	return
}

// Start - Starts reading from queue.
func (crawl *Crawl) Start() (err error) {
	defer func() {
		crawl.doneCh <- true
	}()

	work := make(chan *Request, crawl.Options.QueueConcurrency)

	var workers []chan chan bool
	for index := 0; index < crawl.Options.QueueConcurrency; index++ {
		closeChan := crawl.startWorker(work)
		workers = append(workers, closeChan)
	}

	err = crawl.startLoop(work, workers)
	if err != nil {
		return
	}

	for _, closeChan := range workers {
		done := make(chan bool, 1)
		closeChan <- done
		<-done
	}

	return
}

// Start scheduling requests
// until we have something to do.
// Also wait for new requests if some
// are already executing in background.
func (crawl *Crawl) startLoop(work chan *Request, workers []chan chan bool) (err error) {
	// Ticker for request scheduling
	// it ticks every (second / max request per second)
	// making it schedule maximum (max requests per second)
	// of requests per second
	var tick <-chan time.Time
	if crawl.Options.MaxRequestsPerMinute > 0 {
	} else if crawl.Options.MaxRequestsPerSecond > 0 {
		tick = time.Tick(time.Second / time.Duration(crawl.Options.MaxRequestsPerSecond))
	} else {
		return errors.New("MaxRequestsPerMinute or MaxRequestsPerMinute must be set")
	}

	for crawl.Queue.Continue() {
		if tick != nil {
			<-tick
		}

		// Stop if got a close signal
		select {
		case <-crawl.closeCh:
			return
		default:
		}

		// Get request from queue and execute it
		if request, ok := crawl.Queue.Get(); ok {
			work <- request
		} else {
			return
		}
	}
	return
}

func (crawl *Crawl) startWorker(work chan *Request) (closeChan chan chan bool) {
	closeChan = make(chan chan bool, 1)
	go func() {
		for {
			select {
			case req := <-work:
				crawl.Execute(req)
				crawl.Queue.Done()
			case done := <-closeChan:
				done <- true
				return
			}
		}
	}()
	return
}

// Schedule - Schedules request for future execution.
// Request will be executed as soon as possible.
// Execution of requests is limited by MaxRequestsPerSecond.
func (crawl *Crawl) Schedule(req *Request) {
	crawl.Queue.Schedule(req)
}

// Handler - Adds new crawl handler.
// Handler is a callback referenced by name.
func (crawl *Crawl) Handler(name interface{}, h Handler) {
	crawl.mutex.Lock()
	crawl.handlers[name] = append(crawl.handlers[name], h)
	crawl.mutex.Unlock()
}

// GetHandlers - Gets crawl handlers by name.
func (crawl *Crawl) GetHandlers(name interface{}) []Handler {
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
