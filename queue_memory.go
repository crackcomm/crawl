package crawl

import (
	"io"
	"sync"

	"golang.org/x/net/context"
)

// NewQueue - Makes a new queue.
// Capacity argument is a capacity of requests channel.
func NewQueue(capacity int) Queue {
	jobs := make(chan Job, capacity)
	return &memQueue{
		writeChan: jobs,
		readChan:  jobs,
		mutex:     new(sync.RWMutex),
	}
}

type memQueue struct {
	writeChan chan Job
	readChan  chan Job
	mutex     *sync.RWMutex
}

func (queue *memQueue) Get() (Job, error) {
	job, ok := <-queue.readChan
	if !ok {
		return nil, io.EOF
	}
	return job, nil
}

func (queue *memQueue) Schedule(ctx context.Context, r *Request) error {
	// We are locking the mutex so we can avoid writes to closed channel.
	// Close will change writeChan to nil after closing it and requests
	// will be drained from readChan (still the same but closed channel).
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()
	if queue.writeChan == nil {
		return io.ErrClosedPipe
	}
	queue.writeChan <- &memJob{ctx: ctx, req: r}
	return nil
}

func (queue *memQueue) Close() (err error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.writeChan == nil {
		return
	}
	channel := queue.writeChan
	queue.writeChan = nil
	close(channel)
	return
}

// memJob - Structure to make Request+Context a Job interface.
type memJob struct {
	req *Request
	ctx context.Context
}

func (job *memJob) Context() context.Context {
	return job.ctx
}

func (job *memJob) Request() *Request {
	return job.req
}

func (job *memJob) Done() {
}
