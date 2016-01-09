package nsqcrawl

import (
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/crackcomm/crawl"
	"github.com/crackcomm/nsqueue/consumer"
	"github.com/crackcomm/nsqueue/producer"
)

// NewQueue - Creates nsq consumer and producer.
func NewQueue(topic, channel string, maxInFlight int) *Queue {
	q := &Queue{
		Consumer: consumer.New(),
		Producer: producer.New(),
		memq:     crawl.NewQueue(maxInFlight + 1),
		topic:    topic,
	}
	q.Consumer.Register(topic, channel, maxInFlight, q.nsqHandler)
	return q
}

// Queue - NSQ Queue.
type Queue struct {
	*consumer.Consumer
	*producer.Producer

	topic string
	memq  crawl.Queue
}

// Get - Gets job from channel.
func (queue *Queue) Get() (job crawl.Job, err error) {
	return queue.memq.Get()
}

// Schedule - Schedules job in nsq.
// It will not call job.Done ever.
func (queue *Queue) Schedule(job crawl.Job) (err error) {
	r := &request{Request: job.Request()}
	if deadline, ok := job.Context().Deadline(); ok {
		r.Deadline = deadline
	}
	return queue.Producer.PublishJSON(queue.topic, r)
}

// Close - Closes consumer and producer.
func (queue *Queue) Close() (err error) {
	queue.Producer.Stop()
	queue.Consumer.Stop()
	return queue.memq.Close()
}

func (queue *Queue) nsqHandler(msg *consumer.Message) {
	req := new(request)
	err := msg.ReadJSON(req)
	if err != nil {
		glog.V(3).Infof("nsq json (%s) error: %v", msg.Body, err)
		msg.GiveUp()
		return
	}

	// Check if deadline exceeded
	if !req.Deadline.IsZero() && time.Now().After(req.Deadline) {
		glog.V(3).Infof("request deadline exceeded (%s)", msg.Body)
		msg.GiveUp()
		return
	}

	// Request context
	ctx := context.Background()

	// Set request deadline
	if !req.Deadline.IsZero() {
		ctx, _ = context.WithDeadline(ctx, req.Deadline)
	}

	// Set nsq message in context
	ctx = consumer.WithMessage(ctx, msg)

	// Schedule job in memory
	queue.memq.Schedule(&nsqJob{msg: msg, req: req.Request, ctx: ctx})
}

type request struct {
	Request  *crawl.Request `json:"request,omitempty"`
	Deadline time.Time      `json:"deadline,omitempty"`
}

type nsqJob struct {
	msg *consumer.Message
	req *crawl.Request
	ctx context.Context
}

func (job *nsqJob) Context() context.Context { return job.ctx }
func (job *nsqJob) Request() *crawl.Request  { return job.req }
func (job *nsqJob) Done()                    { job.msg.Success() }
