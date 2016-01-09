package crawl

import (
	"net/http"
	"time"
)

// Option - Crawl option.
type Option func(*crawl)

// options - Crawl options.
type options struct {
	concurrency   int
	queueCapacity int
	headers       map[string]string

	defaultTimeout time.Duration
}

// WithTransport - Sets crawl HTTP transport.
func WithTransport(transport *http.Transport) Option {
	return func(c *crawl) {
		c.transport = transport
	}
}

// WithQueue - Sets crawl queue.
// Default: creates queue using NewQueue() with capacity of WitWithQueueCapacity().
func WithQueue(queue Queue) Option {
	return func(c *crawl) {
		c.queue = queue
	}
}

// WithDefaultHeaders - Sets crawl default headers.
// Default: empty.
func WithDefaultHeaders(headers map[string]string) Option {
	return func(c *crawl) {
		c.opts.headers = headers
	}
}

// WithUserAgent - Sets crawl default user-agent.
func WithUserAgent(ua string) Option {
	return func(c *crawl) {
		if c.opts.headers == nil {
			c.opts.headers = make(map[string]string)
		}
		c.opts.headers["User-Agent"] = ua
	}
}

// WithConcurrency - Sets crawl concurrency.
// Default: 1000.
func WithConcurrency(n int) Option {
	return func(c *crawl) {
		c.opts.concurrency = n
	}
}

// WithQueueCapacity - Sets queue capacity.
// It sets queue capacity if a queue needs to be created and it sets a capacity of channel in-memory queue.
// It also sets capacity of errors buffered channel.
// Default: 10000.
func WithQueueCapacity(n int) Option {
	return func(c *crawl) {
		c.opts.queueCapacity = n
	}
}

// WithSpiders - Registers spider on a crawler.
func WithSpiders(spiders ...func(Crawler)) Option {
	return func(c *crawl) {
		for _, spider := range spiders {
			spider(c)
		}
	}
}

// WithDefaultTimeout - Sets default request timeout duration.
func WithDefaultTimeout(d time.Duration) Option {
	return func(c *crawl) {
		c.opts.defaultTimeout = d
	}
}
