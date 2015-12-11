package crawl

import (
	"errors"
	"sync/atomic"
)

// memQueue - In-memory requests queue.
type memQueue struct {
	processing int64
	done       int64

	ch chan *Request
}

// ErrQueueClosed - Error returned when tried to pull a request from closed queue.
var ErrQueueClosed = errors.New("Queue closed")

// NewQueue - Makes a new queue.
// Capacity argument is a capacity of requests channel.
func NewQueue(capacity int) Queue {
	return &memQueue{
		ch: make(chan *Request, capacity),
	}
}

// Schedule - Schedules Request execution in Queue.
// Pushes request to Queue channel.
func (queue *memQueue) Schedule(req *Request) error {
	queue.ch <- req
	return nil
}

// Get - Gets request from Queue channel.
// If queue is freezed waits until unfreezed.
// OK is true if request is not nil.
// Processing counter is incremented when request exits the queue.
// It should be then decremented by using Done() method.
func (queue *memQueue) Get() (req *Request, err error) {
	// Get req from channel
	req = <-queue.ch
	if req == nil {
		return nil, ErrQueueClosed
	}

	atomic.AddInt64(&queue.processing, 1)
	return
}

// Continue - Returns true if there is at least one job in progress
// or there is something in the queue. Otherwise returns false.
func (queue *memQueue) Continue() bool {
	return queue.processing > 0 || len(queue.ch) > 0
}

// Done - Increments done counter and decrements processing counter.
// Closes queue if Continue() returns false.
func (queue *memQueue) Done() {
	// Add one to done counter
	// and decrease processing counter
	atomic.AddInt64(&queue.processing, -1)
	atomic.AddInt64(&queue.done, 1)

	// Return if queue is ok
	// And there is still something to do.
	if queue.Continue() {
		return
	}

	// If there are no requests in progress
	// and to requests to proceed - queue is closed.
	queue.Close()
}

// Close - Closes queue.
func (queue *memQueue) Close() error {
	close(queue.ch)
	return nil
}
