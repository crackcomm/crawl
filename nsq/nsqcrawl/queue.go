package nsqcrawl

import (
	"io"
	"time"

	"google.golang.org/grpc/metadata"

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
		channel:  make(chan *nsqJob, maxInFlight+1),
		topic:    topic,
	}
	q.Consumer.Register(topic, channel, maxInFlight, q.nsqHandler)
	return q
}

// NewProducer - Creates queue producer.
func NewProducer(topic string) *Queue {
	return &Queue{
		Producer: producer.New(),
		topic:    topic,
	}
}

// Queue - NSQ Queue.
type Queue struct {
	*consumer.Consumer
	*producer.Producer

	topic   string
	channel chan *nsqJob
}

// Schedule - Schedules job in nsq.
// It will not call job.Done ever.
func (queue *Queue) Schedule(ctx context.Context, req *crawl.Request) (err error) {
	md, _ := metadata.FromContext(ctx)
	r := &Request{Request: req, Metadata: md}
	if deadline, ok := ctx.Deadline(); ok {
		r.Deadline = deadline
	}
	return queue.Producer.PublishJSON(queue.topic, r)
}

// Get - Gets job from channel.
func (queue *Queue) Get() (crawl.Job, error) {
	job, ok := <-queue.channel
	if !ok {
		return nil, io.EOF
	}
	return job, nil
}

// Close - Closes consumer and producer.
func (queue *Queue) Close() (err error) {
	if queue.Producer != nil {
		queue.Producer.Stop()
	}
	if queue.Consumer != nil {
		queue.Consumer.Stop()
	}
	if queue.channel != nil {
		close(queue.channel)
	}
	return
}

func (queue *Queue) nsqHandler(msg *consumer.Message) {
	req := new(Request)
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

	// Set metadata in context
	if len(req.Metadata) > 0 {
		ctx = metadata.NewContext(ctx, req.Metadata)
	}

	// Schedule job in memory
	queue.channel <- &nsqJob{msg: msg, req: req.Request, ctx: ctx}
}

// Request - Request as it is in NSQ.
type Request struct {
	Request  *crawl.Request `json:"request,omitempty"`
	Deadline time.Time      `json:"deadline,omitempty"`
	Metadata metadata.MD    `json:"metadata,omitempty"`
}

type nsqJob struct {
	msg *consumer.Message
	req *crawl.Request
	ctx context.Context
}

func (job *nsqJob) Context() context.Context { return job.ctx }
func (job *nsqJob) Request() *crawl.Request  { return job.req }
func (job *nsqJob) Done()                    { job.msg.Success() }
