package crawl

// Queue - Requests queue.
type Queue interface {
	// Schedule - Schedules a Request.
	Schedule(req *Request) error

	// Continue - Returns true if there is at least one job in progress
	// or there is something in the queue. Otherwise returns false.
	// It may return true always for example when it's a NSQ queue.
	Continue() bool

	// Get - Gets request from Queue channel.
	// Processing counter is incremented when request exits the queue.
	// It should be then decremented by using Done() method.
	Get() (req *Request, err error)

	// Done - Increments done counter and decrements processing counter.
	// Closes queue if Continue() returns false.
	Done()

	// Close - Closes the queue.
	Close() error
}
