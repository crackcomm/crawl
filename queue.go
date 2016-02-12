package crawl

import "golang.org/x/net/context"

// Job - Crawl job interface.
type Job interface {
	// Request - Returns crawl job.
	Request() *Request

	// Context - Returns job context.
	Context() context.Context

	// Done - Sets job as done.
	Done()
}

// Queue - Requests queue.
type Queue interface {
	// Get - Gets request from Queue channel.
	// Returns io.EOF if queue is empty.
	Get() (Job, error)

	// Schedule - Schedules a Request.
	// Returns io.ErrClosedPipe if queue is closed.
	Schedule(context.Context, *Request) error

	// Close - Closes the queue.
	Close() error
}
