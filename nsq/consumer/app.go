// Package consumer implements command line crawl consumer from nsq.
package consumer

import (
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"gopkg.in/urfave/cli.v2"

	clinsq "github.com/crackcomm/cli-nsq"
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
	// using parameters from command line.
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
	clinsq.AddrFlag,
	clinsq.LookupAddrFlag,
	clinsq.TopicFlag,
	clinsq.ChannelFlag,
	&cli.IntFlag{
		Name:    "concurrency",
		Value:   100,
		EnvVars: []string{"CONCURRENCY"},
	},
	&cli.IntFlag{
		Name:    "timeout",
		Usage:   "default timeout in seconds",
		Value:   30,
		EnvVars: []string{"TIMEOUT"},
	},
}

// New - Creates nsq consumer app.
func New(opts ...Option) *cli.App {
	app := &App{opts: opts}
	cliapp := (&cli.App{})
	cliapp.Name = "crawler"
	cliapp.HelpName = cliapp.Name
	cliapp.Version = "0.0.1"
	cliapp.Usage = "nsq crawl consumer"
	cliapp.Flags = Flags
	cliapp.Action = app.Action
	return cliapp
}

// Action - Command line action.
func (app *App) Action(c *cli.Context) error {
	app.Ctx = c
	app.Queue = nsqcrawl.NewQueue(c.String("topic"), c.String("channel"), c.Int("concurrency"))

	for _, opt := range app.opts {
		opt(app)
	}

	if err := app.Before(c); err != nil {
		return err
	}

	// Construct crawler
	// It has to be done before connecting to nsq
	crawler := app.Crawler()

	// Register all spiders
	for _, spiderConstructor := range app.spiderConstructors {
		if spider := spiderConstructor(app); spider != nil {
			spider(crawler)
		}
	}

	// Connect to nsq and nsqlookup
	if err := clinsq.Connect(c); err != nil {
		return err
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
			return nil
		case s := <-sig:
			glog.Infof("Received signal %v, closing crawler", s)
			return app.Queue.Close()
		}
	}
}

// Before - Executed before action.
func (app *App) Before(c *cli.Context) (err error) {
	if app.before != nil {
		err = app.before(app)
		if err != nil {
			return
		}
	}
	return clinsq.RequireAll(c)
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
		crawl.WithDefaultTimeout(time.Duration(app.Ctx.Int("timeout"))*time.Second),
	)
}
