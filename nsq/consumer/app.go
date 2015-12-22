// Package consumer implements command line crawl consumer from nsq.
package consumer

import (
	"errors"

	"github.com/codegangsta/cli"
	"github.com/golang/glog"

	"github.com/crackcomm/crawl"
)

// App - Consumer command line application structure.
type App struct {
	// Ctx - Cli context, set on action.
	Ctx *cli.Context

	// NsqQueue - NSQ queue. Constructed as first during Action() call.
	*crawl.NsqQueue

	// before - Flag requirements checking.
	before func(c *cli.Context) error

	// crawler - Accessed using Crawler() which constructs it on first call
	// using parameters from commmand line.
	crawler crawl.Crawler

	// crawlerConstructor - Constructs a crawler. Called only once in Crawler().
	// It can be changed using WithCrawlerConstructor()
	crawlerConstructor func(*App) crawl.Crawler

	// opts - Options which are applied on Action() call with all required
	// parameters from the command line context.
	opts []Option
}

// Flags - Consumer app flags.
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:   "topic",
		EnvVar: "TOPIC",
	},
	cli.StringFlag{
		Name:   "channel",
		EnvVar: "CHANNEL",
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
	cliapp.Before = app.Before
	cliapp.Action = app.Action
	return cliapp
}

// Action - Command line action.
func (app *App) Action(c *cli.Context) {
	app.NsqQueue = crawl.NewNsqQueue(c.String("topic"), c.String("channel"), c.Int("concurrency"))

	for _, opt := range app.opts {
		opt(app)
	}

	crawler := app.Crawler()

	nsqAddr := c.StringSlice("nsq-addr")[0]
	if err := app.NsqQueue.Producer.Connect(nsqAddr); err != nil {
		glog.Fatalf("Error connecting producer to %q: %v", nsqAddr, err)
	}

	if addrs := c.StringSlice("nsq-addr"); len(addrs) != 0 {
		for _, addr := range addrs {
			glog.V(3).Infof("Connecting to nsq %s", addr)
			if err := app.NsqQueue.Consumer.Connect(addr); err != nil {
				glog.Fatalf("Error connecting to nsq %q: %v", addr, err)
			}
			glog.V(3).Infof("Connected to nsq %s", addr)
		}
	}

	if addrs := c.StringSlice("nsqlookup-addr"); len(addrs) != 0 {
		for _, addr := range addrs {
			glog.V(3).Infof("Connecting to nsq lookup %s", addr)
			if err := app.NsqQueue.Consumer.ConnectLookupd(addr); err != nil {
				glog.Fatalf("Error connecting to nsq lookup %q: %v", addr, err)
			}
			glog.V(3).Infof("Connected to nsq lookup %s", addr)
		}
	}

	go func() {
		for err := range crawler.Errors() {
			glog.Warningf("Crawl error: %v", err)
		}
	}()

	crawler.Start()
}

// Before - Executed before action.
func (app *App) Before(c *cli.Context) error {
	// Set application context
	app.Ctx = c

	// Use customized before if any
	if app.before != nil {
		return app.before(c)
	}
	return beforeApp(c)
}

func beforeApp(c *cli.Context) error {
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
		crawl.WithQueue(app.NsqQueue),
		crawl.WithConcurrency(app.Ctx.Int("concurrency")),
	)
}
