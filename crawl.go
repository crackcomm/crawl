package crawl

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"golang.org/x/net/context"
)

// Crawl - Crawl structure.
// It keeps track of one-client crawl.
// Requests are scheduled in queue (look QueueCapacity option).
// Then depending on MaxRequestsPerSecond options they are executed.
// Each request should contain at least one Callback.
// Which is a string key to Handler.
//
// HTTP Client proxy can be set in environment variable:
//
// HTTP_PROXY - for http requests,
// HTTPS_PROXY - for https requests,
// NO_PROXY - for no proxy at all
//
// Crawl can be Freeze()-ed when it's required.
type Crawl struct {
	Errors chan *Error // this channel should be emptied by You
	Queue
	*Freezer
	*http.Client

	handlers map[string][]Handler

	closeCh chan bool // close channel
	doneCh  chan bool // done channel

	headers map[string]string // Default HTTP headers
	opts    *options
}

// Error - Crawl error.
type Error struct {
	*Request
	Error error
}

// Handler - Crawl handler.
type Handler func(context.Context, *Response) error

// New - Creates new crawl.
// Options are set permamently on Queue (QueueCapacity).
// When it's changed Queue should be remade.
// By default crawl is created with DefaultOptions
// If you want to change options set crawl.SetOptions() method.
// Only one opts is accepted, rest is ignored.
func New(opts ...Option) (crawl *Crawl) {
	client := &http.Client{
		Transport: http.DefaultTransport,
	}
	client.Jar, _ = cookiejar.New(nil)

	crawl = &Crawl{
		Freezer:  NewFreezer(1000),
		Errors:   make(chan *Error, 1000),
		Client:   client,
		handlers: make(map[string][]Handler),
		closeCh:  make(chan bool, 1),
		doneCh:   make(chan bool, 1),
		opts:     &options{concurrency: 1},
	}
	for _, opt := range opts {
		opt(crawl)
	}
	return
}

// Do - Makes http.Request using crawl.Client
// returns http.Response wrapped in Response structure.
func (crawl *Crawl) Do(req *Request) (resp *Response, err error) {
	// Get http.Request structure
	httpreq, err := req.HTTPRequest()
	if err != nil {
		return
	}

	// Copy default headers
	for name, value := range crawl.headers {
		if _, has := httpreq.Header[name]; !has {
			httpreq.Header.Set(name, value)
		}
	}

	// Send request and read response
	response, err := crawl.Client.Do(httpreq)
	if err != nil {
		return
	}

	return &Response{Response: response}, nil
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

	work := make(chan Job, crawl.opts.concurrency)

	var workers []chan chan bool
	for index := 0; index < crawl.opts.concurrency; index++ {
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
func (crawl *Crawl) startLoop(work chan Job, workers []chan chan bool) (err error) {
	// Ticker for request scheduling
	// it ticks every (second / max request per second)
	// making it schedule maximum (max requests per second)
	// of requests per second
	var tick <-chan time.Time
	if crawl.opts.maxReqsPerSecond > 0 {
		tick = time.Tick(time.Second / time.Duration(crawl.opts.maxReqsPerSecond))
	}

	for {
		if crawl.opts.maxReqsPerSecond > 0 {
			<-tick
		}

		// Stop if got a close signal
		select {
		case <-crawl.closeCh:
			return
		default:
		}

		// Wait if freezed
		crawl.Freezer.Wait()

		// Get request from queue and execute it
		if job, err := crawl.Queue.Get(); err == nil {
			work <- job
		} else if err == io.EOF {
			return nil
		} else {
			return err
		}
	}
}

func (crawl *Crawl) startWorker(work chan Job) (closeChan chan chan bool) {
	closeChan = make(chan chan bool, 1)
	go func() {
		for {
			select {
			case job := <-work:
				if _, err := crawl.Execute(job.Request()); err != nil {
					crawl.Errors <- &Error{Error: err, Request: job.Request()}
				}
				job.Done()
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
func (crawl *Crawl) Schedule(job Job) {
	crawl.Queue.Schedule(job)
}

// Handler - Adds new crawl handler.
// Handler is a callback referenced by name.
func (crawl *Crawl) Handler(name string, h Handler) {
	crawl.handlers[name] = append(crawl.handlers[name], h)
}

// GetHandlers - Gets crawl handlers by name.
func (crawl *Crawl) GetHandlers(name string) []Handler {
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
	close(crawl.Errors)
	return
}
