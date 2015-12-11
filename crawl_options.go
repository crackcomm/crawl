package crawl

import "net/http"

// Option - Crawl option.
type Option func(*Crawl)

// options - Crawl options.
type options struct {
	concurrency      int
	maxReqsPerSecond int
}

// WithQueue - Sets crawl queue.
func WithQueue(queue Queue) Option {
	return func(crawl *Crawl) {
		crawl.Queue = queue
	}
}

// WithConcurrency - Sets crawl concurrency.
func WithConcurrency(n int) Option {
	return func(crawl *Crawl) {
		crawl.opts.concurrency = n
	}
}

// WithMaxReqsPerSec - Sets crawl maximum amount of requests made per second.
func WithMaxReqsPerSec(n int) Option {
	return func(crawl *Crawl) {
		crawl.opts.maxReqsPerSecond = n
	}
}

// WithClient - Sets crawl HTTP client.
func WithClient(client *http.Client) Option {
	return func(crawl *Crawl) {
		crawl.Client = client
	}
}
