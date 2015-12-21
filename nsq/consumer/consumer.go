// Package consumer implements command line crawl consumer from nsq.
package consumer

import (
	"github.com/codegangsta/cli"
	"github.com/golang/glog"

	"github.com/crackcomm/crawl"
)

// Spider - Spider registrator.
type Spider func(crawl.Crawler)

// New - Creates nsq consumer app.
func New(spiders ...Spider) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "crawler"
	app.HelpName = app.Name
	app.Version = "0.0.1"
	app.Usage = "nsq crawl consumer"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "requests-topic",
			Value:  "crawl_requests",
			EnvVar: "REQUESTS_TOPIC",
		},
		cli.StringSliceFlag{
			Name:   "nsqlookup-addr",
			EnvVar: "NSQLOOKUP_ADDR",
		},
		cli.StringSliceFlag{
			Name:   "nsq-addr",
			EnvVar: "NSQ_ADDR",
		},
		cli.StringFlag{
			Name:   "consumer-channel",
			Value:  "consumer",
			EnvVar: "CONSUMER_CHANNEL",
		},
		cli.StringFlag{
			Name:   "producer-topic",
			Value:  "crawl_results",
			EnvVar: "PRODUCER_TOPIC",
		},
		cli.IntFlag{
			Name:   "concurrency",
			Value:  100,
			EnvVar: "CRAWL_CONCURRENCY",
		},
	}
	app.Action = func(c *cli.Context) {
		if len(c.StringSlice("nsq-addr")) == 0 {
			glog.Fatalf("At leat one --%s is required", "nsq-addr")
		}

		queue := crawl.NewNsqQueue(c.String("requests-topic"), c.String("consumer-channel"), c.Int("concurrency"))

		nsqAddr := c.StringSlice("nsq-addr")[0]
		if err := queue.Producer.Connect(nsqAddr); err != nil {
			glog.Fatalf("Error connecting producer to %q: %v", nsqAddr, err)
		}

		if addrs := c.StringSlice("nsq-addr"); len(addrs) != 0 {
			for _, addr := range addrs {
				glog.V(3).Infof("Connecting to nsq %s", addr)
				if err := queue.Consumer.Connect(addr); err != nil {
					glog.Fatalf("Error connecting to nsq %q: %v", addr, err)
				}
				glog.V(3).Infof("Connected to nsq %s", addr)
			}
		}

		if addrs := c.StringSlice("nsqlookup-addr"); len(addrs) != 0 {
			for _, addr := range addrs {
				glog.V(3).Infof("Connecting to nsq lookup %s", addr)
				if err := queue.Consumer.ConnectLookupd(addr); err != nil {
					glog.Fatalf("Error connecting to nsq lookup %q: %v", addr, err)
				}
				glog.V(3).Infof("Connected to nsq lookup %s", addr)
			}
		}

		crawler := crawl.New(
			crawl.WithQueue(queue),
			crawl.WithConcurrency(c.Int("concurrency")),
		)

		for _, spider := range spiders {
			spider(crawler)
		}

		go func() {
			for err := range crawler.Errors() {
				glog.Warningf("Crawl error: %v", err)
			}
		}()

		crawler.Start()
	}
	return
}
