package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	cliflags "github.com/crackcomm/cli-flags"
	"github.com/crackcomm/cli-nsq"
	"github.com/crackcomm/crawl"
	"github.com/crackcomm/crawl/nsq/nsqcrawl"
	"github.com/golang/glog"
	"github.com/nsqio/go-nsq"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	defer glog.Flush()

	// CRAWL_DEBUG environment variable turns on debug mode
	// crawler then can spit out logs using glog.V(3)
	var verbosity string
	if yes, _ := strconv.ParseBool(os.Getenv("CRAWL_DEBUG")); yes {
		verbosity = "-v=3"
	}

	// We are setting glog to log to stderr
	flag.CommandLine.Parse([]string{"-logtostderr", verbosity})

	app := (&cli.App{})
	app.Name = "crawl-schedule"
	app.ArgsUsage = "<url>"
	app.HelpName = app.Name
	app.Version = "0.0.1"
	app.Usage = "schedules a crawl request in nsq"
	app.Flags = []cli.Flag{
		clinsq.AddrFlag,
		clinsq.TopicFlag,
		&cli.StringSliceFlag{
			Name:  "form-value",
			Usage: "form value in format (format: key=value)",
		},
		&cli.StringSliceFlag{
			Name:  "metadata",
			Usage: "metadata value in format (format: key=value)",
		},
		&cli.StringSliceFlag{
			Name:  "callback",
			Usage: "crawl request callbacks (required)",
		},
		&cli.StringFlag{
			Name:  "referer",
			Usage: "crawl request referer",
		},
		&cli.StringFlag{
			Name:  "method",
			Usage: "crawl request referer",
			Value: "GET",
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "request timeout",
		},
	}
	app.Before = func(c *cli.Context) error {
		if err := cliflags.RequireAll(c, []cli.Flag{
			clinsq.AddrFlag,
			clinsq.TopicFlag,
		}); err != nil {
			return err
		}
		if len(c.StringSlice("callback")) == 0 {
			return errors.New("--callback flag is missing")
		}
		if c.Args().Len() != 1 {
			return errors.New("URL argument is missing")
		}
		return nil
	}
	app.Action = func(c *cli.Context) error {
		form, err := listToForm(c.StringSlice("form-value"))
		if err != nil {
			return fmt.Errorf("Form values error: %v", err)
		}
		md, err := listToForm(c.StringSlice("metadata"))
		if err != nil {
			return fmt.Errorf("Metadata values error: %v", err)
		}

		request := &crawl.Request{
			URL:       strings.Trim(c.Args().First(), `"'`),
			Form:      form,
			Method:    c.String("method"),
			Referer:   c.String("referer"),
			Callbacks: c.StringSlice("callback"),
		}

		ctx := context.Background()
		if len(md) > 0 {
			ctx = metadata.NewContext(ctx, metadata.MD(md))
		}

		if glog.V(3) {
			body, _ := json.MarshalIndent(request, "", "  ")
			glog.Infof("Scheduling request: %s", body)
		}

		// Set context deadline
		if timeout := c.Duration("timeout"); timeout > 0 {
			ctx, _ = context.WithTimeout(ctx, timeout)
		}

		// Create nsq queue
		q := nsqcrawl.NewProducer(c.String("topic"))
		defer q.Close()

		// Connect to nsq
		cfg := nsq.NewConfig()
		cfg.OutputBufferTimeout = 0
		if err := q.Producer.ConnectConfig(c.String("nsq-addr"), cfg); err != nil {
			return fmt.Errorf("Error connecting to nsq: %v", err)
		}

		// Configure NSQ producer logger
		q.Producer.SetLogger(log.New(os.Stdout, "[nsq]", 0), nsq.LogLevelError)

		// Schedule request
		if err := q.Schedule(ctx, request); err != nil {
			return fmt.Errorf("schedule error: %v", err)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		glog.Fatal(err)
	}
}

func listToForm(list []string) (result url.Values, err error) {
	result = make(url.Values)
	for _, keyValue := range list {
		i := strings.Index(keyValue, "=")
		if i <= 0 {
			return nil, fmt.Errorf("%q is not valid", keyValue)
		}
		key := keyValue[:i]
		value := keyValue[i+1:]
		result.Set(key, value)
	}
	return
}

func mapStringsToInterfaces(input map[string]string) (result map[string]interface{}) {
	result = make(map[string]interface{})
	for key, value := range input {
		result[key] = value
	}
	return
}
