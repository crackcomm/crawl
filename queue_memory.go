package crawl

import "io"

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

func (queue *memQueue) Schedule(job Job) error {
	if queue.channel == nil {
		return io.ErrClosedPipe
	}
	queue.channel <- job
	return nil
}

func (queue *memQueue) Close() error {
	ch := queue.channel
	queue.channel = nil
	close(ch)
	return nil
}
