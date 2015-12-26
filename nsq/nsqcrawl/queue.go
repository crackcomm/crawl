package nsqcrawl

import (
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
	return queue.Producer.PublishJSON(queue.topic, job.Request())
}

// Close - Closes consumer and producer.
func (queue *Queue) Close() (err error) {
	queue.Producer.Stop()
	queue.Consumer.Stop()
	return queue.memq.Close()
}

func (queue *Queue) nsqHandler(msg *consumer.Message) {
	req := new(crawl.Request)
	err := msg.ReadJSON(req)
	glog.V(3).Infof("nsq json: %s", msg.Body)
	if err != nil {
		glog.V(3).Infof("nsq json (%s) error: %v", msg.Body, err)
		msg.GiveUp()
		return
	}

	// Set nsq message in context
	ctx := consumer.WithMessage(context.Background(), msg)

	// Schedule job in memory
	queue.memq.Schedule(&nsqJob{msg: msg, req: req, ctx: ctx})
}

type nsqJob struct {
	msg *consumer.Message
	req *crawl.Request
	ctx context.Context
}

func (job *nsqJob) Context() context.Context { return job.ctx }
func (job *nsqJob) Request() *crawl.Request  { return job.req }
func (job *nsqJob) Done()                    { job.msg.Success() }
