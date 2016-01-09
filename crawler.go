package crawl

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/ryanuber/go-glob"

	"golang.org/x/net/context"
)

// Handler - Crawler handler.
type Handler func(context.Context, *Response) error

// Middleware - Crawler middleware.
type Middleware func(context.Context, *Request, *http.Request) error

// Crawler - Crawler interface.
type Crawler interface {
	// Schedule - Schedules request.
	// Context is passed to queue in a job.
	Schedule(context.Context, *Request) error

	// Execute - Makes a http request respecting context deadline.
	// If request Raw is not true - ParseHTML() method is executed on Response.
	// Then all callbacks are executed with context.
	Execute(context.Context, *Request) (*Response, error)

	// Handlers - Returns all registered handlers.
	Handlers() map[string][]Handler

	// Register - Registers crawl handler.
	Register(name string, h Handler)

	// Middleware - Registers a middleware.
	// Request is not executed if middleware returns an error.
	Middleware(Middleware)

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
		handlers:   make(map[string][]Handler),
		opts:       &options{concurrency: 1000, queueCapacity: 10000, headers: DefaultHeaders},
		errorsChan: make(chan error, 10000),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.transport == nil {
		c.transport = new(http.Transport)
	}
	c.client = &http.Client{Transport: c.transport}
	if c.client.Jar == nil {
		c.client.Jar, _ = cookiejar.New(nil)
	}
	if c.queue == nil {
		c.queue = NewQueue(c.opts.queueCapacity)
	}
	return c
}

// DefaultHeaders - Default crawler headers.
var DefaultHeaders = map[string]string{
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"Accept-Language": "en-US,en;q=0.8",
	"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36",
}

// crawl - Crawler structure.
type crawl struct {
	errorsChan chan error
	handlers   map[string][]Handler
	transport  *http.Transport
	client     *http.Client
	opts       *options

	queue Queue

	// patterns - callbacks glob patterns
	patterns []string

	// middlewares - crawler middlewares.
	middlewares []Middleware
}

func (crawl *crawl) Start() {
	wg := new(sync.WaitGroup)
	for i := 0; i < crawl.opts.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
	// Set default timeout if enabled and empty in context
	if _, ok := ctx.Deadline(); !ok && crawl.opts.defaultTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(crawl.opts.defaultTimeout))
		defer cancel()
	}

	// Get http.Request structure
	httpReq, err := ConstructHTTPRequest(req)
	if err != nil {
		return
	}

	// Run request middlewares
	for _, middleware := range crawl.middlewares {
		if err = middleware(ctx, req, httpReq); err != nil {
			return
		}
	}

	// Copy default headers
	for name, value := range crawl.opts.headers {
		if _, has := httpReq.Header[name]; !has {
			httpReq.Header.Set(name, value)
		}
	}

	// Send request and read response
	err = crawl.httpDo(ctx, httpReq, func(httpResp *http.Response) error {
		resp = &Response{
			Response: httpResp,
			Request:  req,
		}
		return nil
	})
	if err != nil {
		return
	}
	defer resp.Close()

	// Parse HTML if not request.Raw
	if !req.Raw {
		err = resp.ParseHTML()
		if err != nil {
			return
		}
	}

	if err = crawl.executeHandlers(ctx, resp); err != nil {
		return nil, err
	}

	return
}

func (crawl *crawl) executeHandlers(ctx context.Context, resp *Response) (err error) {
	handlers := crawl.getHandlers(resp.Request.Callbacks)
	if len(handlers) == 0 {
		return
	}
	for _, handler := range handlers {
		if err = handler(ctx, resp); err != nil {
			return
		}
	}
	return
}

func (crawl *crawl) getHandlers(callbacks []string) (list []Handler) {
	for _, pattern := range crawl.patterns {
		for _, name := range callbacks {
			if glob.Glob(pattern, name) {
				list = append(list, crawl.handlers[pattern]...)
				break
			}
		}
	}
	for _, name := range callbacks {
		list = append(list, crawl.handlers[name]...)
	}
	return
}

func (crawl *crawl) Middleware(m Middleware) {
	crawl.middlewares = append(crawl.middlewares, m)
}

func (crawl *crawl) Register(name string, h Handler) {
	if _, ok := crawl.handlers[name]; !ok && strings.Contains(name, "*") {
		crawl.patterns = append(crawl.patterns, name)
	}
	crawl.handlers[name] = append(crawl.handlers[name], h)
}

func (crawl *crawl) Schedule(ctx context.Context, req *Request) error {
	return crawl.queue.Schedule(ctx, req)
}

func (crawl *crawl) Errors() <-chan error {
	return crawl.errorsChan
}

func (crawl *crawl) Handlers() map[string][]Handler {
	return crawl.handlers
}

func (crawl *crawl) httpDo(ctx context.Context, req *http.Request, f func(*http.Response) error) error {
	c := make(chan error, 1)
	go func() {
		resp, err := crawl.client.Do(req)
		if err != nil {
			c <- err
		} else {
			c <- f(resp)
		}
	}()
	select {
	case <-ctx.Done():
		crawl.transport.CancelRequest(req)
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}
