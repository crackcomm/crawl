package crawl

import (
	"sync"
	"sync/atomic"
)

// Queue - Requests queue.
// Queue can be freezed.
// Then it still accepts requests
// but no requests exit the queue.
type Queue struct {
	processing int64
	done       int64

	ch chan *Request

	mutex           *sync.RWMutex
	freezed         bool
	freezeNotifiers chan chan bool
	closed          bool
}

// NewQueue - Makes a new queue.
// Capacity argument is a capacity of requests channel.
func NewQueue(capacity int) *Queue {
	return &Queue{
		ch: make(chan *Request, capacity),

		mutex:           new(sync.RWMutex),
		freezeNotifiers: make(chan chan bool, capacity),
	}
}

// Schedule - Schedules Request execution in Queue.
// Pushes request to Queue channel.
func (queue *Queue) Schedule(req *Request) {
	queue.ch <- req
}

// Get - Gets request from Queue channel.
// If queue is freezed waits until unfreezed.
// OK is true if request is not nil.
// Processing counter is incremented when request exits the queue.
// It should be then decremented by using Done() method.
func (queue *Queue) Get() (req *Request, ok bool) {
	// Wait if freezed
	queue.freezeWait()

	// Get req from channel
	req = <-queue.ch
	if req == nil {
		return nil, false
	}

	atomic.AddInt64(&queue.processing, 1)
	ok = true
	return
}

// Freeze - Freezes the queue.
// Then it doesn't let any requests exit.
// If queue is freezed does nothing.
func (queue *Queue) Freeze() {
	if queue.freezed {
		return
	}

	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.freezed = true
}

// Unfreeze - Unfreezes the queue and frees requests to go.
func (queue *Queue) Unfreeze() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if !queue.freezed {
		return
	}

	queue.freezed = false

	for {
		select {
		case ch := <-queue.freezeNotifiers:
			ch <- true
		default:
			return
		}
	}
}

// Continue - Returns true if there is at least one job in progress
// or there is something in the queue. Otherwise returns false.
func (queue *Queue) Continue() bool {
	return queue.processing > 0 || len(queue.ch) > 0
}

// Done - Increments done counter and decrements processing counter.
// Closes queue if Continue() returns false.
func (queue *Queue) Done() {
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
func (queue *Queue) Close() error {
	if !queue.closed {
		queue.closed = true
		close(queue.freezeNotifiers)
		close(queue.ch)
	}
	return nil
}

// freezeWait - Waits until queue is unfreezed.
// Returns if queue is not freezed
func (queue *Queue) freezeWait() bool {
	if !queue.freezed {
		return true
	}

	// Add freeze notifier
	queue.mutex.Lock()
	ch := make(chan bool, 1)
	queue.freezeNotifiers <- ch
	queue.mutex.Unlock()

	return <-ch
}
