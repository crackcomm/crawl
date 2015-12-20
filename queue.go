package crawl

// Job - Crawl job interface.
type Job interface {
	// Request - Returns crawl job.
	Request() *Request

	// Done - Sets job as done.
	Done()
}

// Queue - Requests queue.
type Queue interface {
	// Get - Gets request from Queue channel.
	// Returns io.EOF if queue is done/closed.
	Get() (Job, error)

	// Schedule - Schedules a Request.
	// Returns io.ErrClosedPipe if queue is closed.
	Schedule(Job) error

	// Close - Closes the queue.
	Close() error
}
