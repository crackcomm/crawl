// Package consumer implements command line crawl consumer from nsq.
package consumer

import (
	"errors"
	"os"
	"os/signal"

	"github.com/codegangsta/cli"
	"github.com/golang/glog"

	"github.com/crackcomm/crawl"
	"github.com/crackcomm/crawl/nsq/nsqcrawl"
)

// Spider - Spider registrator.
type Spider func(crawl.Crawler)

// App - Consumer command line application structure.
type App struct {
	// Ctx - Cli context, set on action.
	Ctx *cli.Context

	// Queue - NSQ queue. Constructed on first Action() call.
	*nsqcrawl.Queue

	// before - Flag requirements checking.
	before func(c *App) error

	// crawler - Accessed using Crawler() which constructs it on first call
	// using parameters from commmand line.
	crawler crawl.Crawler

	// crawlerConstructor - Constructs a crawler. Called only once in Crawler().
	// It can be changed using WithCrawlerConstructor()
	crawlerConstructor func(*App) crawl.Crawler

	// opts - Options which are applied on Action() call with all required
	// parameters from the command line context.
	opts []Option

	// spiderConstructors - List of functions which use flags to construct spider.
	spiderConstructors []func(*App) Spider
}

// Flags - Consumer app flags.
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:   "topic",
		EnvVar: "TOPIC",
		Usage:  "crawl requests nsq topic (required)",
		Value:  "crawl_requests",
	},
	cli.StringFlag{
		Name:   "channel",
		EnvVar: "CHANNEL",
		Usage:  "crawl requests nsq channel (required)",
		Value:  "default",
	},
	cli.StringSliceFlag{
		Name:   "nsq-addr",
		EnvVar: "NSQ_ADDR",
	},
	cli.StringSliceFlag{
		Name:   "nsqlookup-addr",
		EnvVar: "NSQLOOKUP_ADDR",
	},
	cli.IntFlag{
		Name:   "concurrency",
		Value:  100,
		EnvVar: "CONCURRENCY",
	},
}

// New - Creates nsq consumer app.
func New(opts ...Option) *cli.App {
	app := &App{opts: opts}
	cliapp := cli.NewApp()
	cliapp.Name = "crawler"
	cliapp.HelpName = cliapp.Name
	cliapp.Version = "0.0.1"
	cliapp.Usage = "nsq crawl consumer"
	cliapp.Flags = Flags
	cliapp.Action = app.Action
	return cliapp
}

// Action - Command line action.
func (app *App) Action(c *cli.Context) {
	app.Ctx = c
	app.Queue = nsqcrawl.NewQueue(c.String("topic"), c.String("channel"), c.Int("concurrency"))

	for _, opt := range app.opts {
		opt(app)
	}

	if err := app.Before(c); err != nil {
		glog.Fatal(err)
	}

	crawler := app.Crawler()

	nsqAddr := c.StringSlice("nsq-addr")[0]
	if err := app.Queue.Producer.Connect(nsqAddr); err != nil {
		glog.Fatalf("Error connecting producer to %q: %v", nsqAddr, err)
	}

	if addrs := c.StringSlice("nsq-addr"); len(addrs) != 0 {
		for _, addr := range addrs {
			glog.V(3).Infof("Connecting to nsq %s", addr)
			if err := app.Queue.Consumer.Connect(addr); err != nil {
				glog.Fatalf("Error connecting to nsq %q: %v", addr, err)
			}
			glog.V(3).Infof("Connected to nsq %s", addr)
		}
	}

	if addrs := c.StringSlice("nsqlookup-addr"); len(addrs) != 0 {
		for _, addr := range addrs {
			glog.V(3).Infof("Connecting to nsq lookup %s", addr)
			if err := app.Queue.Consumer.ConnectLookupd(addr); err != nil {
				glog.Fatalf("Error connecting to nsq lookup %q: %v", addr, err)
			}
			glog.V(3).Infof("Connected to nsq lookup %s", addr)
		}
	}

	for _, spiderConstructor := range app.spiderConstructors {
		if spider := spiderConstructor(app); spider != nil {
			spider(crawler)
		}
	}

	go func() {
		for err := range crawler.Errors() {
			glog.Warningf("crawl %v", err)
		}
	}()

	done := make(chan bool, 1)
	go func() {
		crawler.Start()
		done <- true
	}()

	glog.Infof("Started crawler (topic=%q)", c.String("topic"))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for {
		select {
		case <-done:
			glog.Info("Crawler closed")
			return
		case s := <-sig:
			glog.Infof("Received signal %v, closing crawler", s)
			if err := app.Queue.Close(); err != nil {
				glog.Fatalf("Error closing queue: %v", err)
			}
			return
		}
	}
}

// Before - Executed before action.
func (app *App) Before(c *cli.Context) (err error) {
	if app.before != nil {
		return app.before(app)
	}
	return beforeApp(c)
}

func beforeApp(c *cli.Context) error {
	if c.String("topic") == "" {
		return errors.New("Flag --topic cannot be empty.")
	}
	if len(c.StringSlice("nsq-addr")) == 0 && len(c.StringSlice("nsqlookup-addr")) == 0 {
		return errors.New("At least one --nsq-addr or --nsqlookup-addr is required")
	}
	return nil
}

// Crawler - Returns app crawler. Constructs if empty.
func (app *App) Crawler() crawl.Crawler {
	if app.crawler == nil {
		if app.crawlerConstructor != nil {
			app.crawler = app.crawlerConstructor(app)
		} else {
			app.crawler = crawlerConstructor(app)
		}
	}
	return app.crawler
}

func crawlerConstructor(app *App) crawl.Crawler {
	return crawl.New(
		crawl.WithQueue(app.Queue),
		crawl.WithConcurrency(app.Ctx.Int("concurrency")),
	)
}
