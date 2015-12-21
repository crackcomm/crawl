package crawl

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"sync"

	"golang.org/x/net/context"
)

// Handler - Crawler handler.
type Handler func(context.Context, *Response) error

// Crawler - Crawler interface.
type Crawler interface {
	// Schedule - Schedules request.
	// If deadline is specified in context it will not timeout execution of the
	// request. It can timeout request scheduling.
	Schedule(context.Context, *Request) error

	// Execute - Makes a http request using crawl.client.
	// If request Raw is not true ParseHTML() method is executed on Response.
	// Then all callbacks are executed with context.
	Execute(context.Context, *Request) (*Response, error)

	// Handlers - Returns all registered handlers.
	Handlers() map[string][]Handler

	// Register - Registers crawl handler.
	Register(name string, h Handler)

	// Start - Starts the crawler.
	// Crawler will work until queue is empty or closed.
	// All errors should be received from Errors() channel.
	Start()

	// Errors - Returns channel that will receive all crawl errors.
	// Only errors from queued requests are here.
	// Not only request errors but also queue errors.
	Errors() <-chan error
}

// New - Creates new crawler. If queue is not provided it will create
// a memory queue with a capacity of WithQueueCapacity seting value (default=10000).
func New(opts ...Option) Crawler {
	c := &crawl{
		handlers: make(map[string][]Handler),
		opts:     &options{concurrency: 1000, queueCapacity: 10000},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.client == nil {
		c.client = &http.Client{
			Transport: http.DefaultTransport,
		}
	}
	if c.client.Jar == nil {
		c.client.Jar, _ = cookiejar.New(nil)
	}
	if c.queue == nil {
		c.queue = NewQueue(c.opts.queueCapacity)
	}
	return c
}

// crawl - Crawler structure.
type crawl struct {
	errorsChan chan error
	handlers   map[string][]Handler
	client     *http.Client
	opts       *options

	queue Queue
}

func (crawl *crawl) Start() {
	wg := new(sync.WaitGroup)
	for i := 0; i < crawl.opts.concurrency; i++ {
		wg.Add(1)
		go func() {
			for {
				job, err := crawl.queue.Get()
				if err == io.EOF {
					return
				} else if err != nil {
					crawl.errorsChan <- err
					return
				}

				if _, err := crawl.Execute(job.Context(), job.Request()); err != nil {
					crawl.errorsChan <- &RequestError{Err: err, Request: job.Request()}
				}

				job.Done()
			}
		}()
	}
	wg.Wait()
	return
}

func (crawl *crawl) Execute(ctx context.Context, req *Request) (resp *Response, err error) {
	// Get http.Request structure
	httpreq, err := req.HTTPRequest()
	if err != nil {
		return
	}

	// Copy default headers
	for name, value := range crawl.opts.headers {
		if _, has := httpreq.Header[name]; !has {
			httpreq.Header.Set(name, value)
		}
	}

	// Send request and read response
	response, err := crawl.client.Do(httpreq)
	if err != nil {
		return
	}

	// Make temporary response structure
	resp = &Response{Response: response}

	// Parse HTML if not request.Raw
	if !req.Raw {
		err = resp.ParseHTML()
		if err != nil {
			return
		}
	}

	// Run handlers
	for _, cb := range req.Callbacks {
		if handlers, ok := crawl.handlers[cb]; ok && len(handlers) >= 1 {
			for _, handler := range handlers {
				err = handler(ctx, resp)
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

func (crawl *crawl) Register(name string, h Handler) {
	crawl.handlers[name] = append(crawl.handlers[name], h)
}

func (crawl *crawl) Schedule(ctx context.Context, req *Request) error {
	return crawl.queue.Schedule(&reqJob{ctx: ctx, req: req})
}

func (crawl *crawl) Errors() <-chan error {
	return crawl.errorsChan
}

func (crawl *crawl) Handlers() map[string][]Handler {
	return crawl.handlers
}
