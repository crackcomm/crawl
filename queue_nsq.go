package crawl

import (
	"golang.org/x/net/context"

	"github.com/crackcomm/nsqueue/consumer"
	"github.com/crackcomm/nsqueue/producer"
)

// NewNsqQueue - Creates nsq consumer and producer.
func NewNsqQueue(topic, channel string, maxInFlight int) *NsqQueue {
	q := &NsqQueue{
		Consumer: consumer.New(),
		Producer: producer.New(),
		memq:     NewQueue(maxInFlight + 1),
		topic:    topic,
	}
	q.Consumer.Register(topic, channel, maxInFlight, q.nsqHandler)
	return q
}

// NsqQueue - NSQ Queue.
type NsqQueue struct {
	*consumer.Consumer
	*producer.Producer

	topic string
	memq  Queue
}

// Get - Gets job from channel.
func (queue *NsqQueue) Get() (job Job, err error) {
	return queue.memq.Get()
}

// Schedule - Schedules job in nsq.
// It will not call job.Done ever.
func (queue *NsqQueue) Schedule(job Job) (err error) {
	return queue.Producer.PublishJSON(queue.topic, job.Request())
}

// Close - Closes consumer and producer.
func (queue *NsqQueue) Close() (err error) {
	queue.Producer.Stop()
	queue.Consumer.Stop()
	return queue.memq.Close()
}

func (queue *NsqQueue) nsqHandler(msg *consumer.Message) {
	req := new(Request)
	err := msg.ReadJSON(req)
	if err != nil {
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
	req *Request
	ctx context.Context
}

func (job *nsqJob) Context() context.Context { return job.ctx }
func (job *nsqJob) Request() *Request        { return job.req }
func (job *nsqJob) Done()                    { job.msg.Success() }
