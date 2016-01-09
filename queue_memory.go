package crawl

import (
	"io"

	"github.com/crackcomm/crawl"
	"golang.org/x/net/context"
)

// NewQueue - Makes a new queue.
// Capacity argument is a capacity of requests channel.
func NewQueue(capacity int) Queue {
	return &memQueue{
		channel: make(chan Job, capacity),
	}
}

type memQueue struct {
	channel chan Job
}

func (queue *memQueue) Get() (Job, error) {
	job, ok := <-queue.channel
	if !ok {
		return nil, io.EOF
	}
	return job, nil
}

func (queue *memQueue) Schedule(ctx context.Context, r *crawl.Request) error {
	if queue.channel == nil {
		return io.ErrClosedPipe
	}
	queue.channel <- &memJob{ctx: ctx, req: r}
	return nil
}

func (queue *memQueue) Close() error {
	ch := queue.channel
	queue.channel = nil
	close(ch)
	return nil
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
