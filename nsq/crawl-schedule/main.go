package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/crackcomm/crawl"
	"github.com/golang/glog"
	"github.com/nsqio/go-nsq"
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

	app := cli.NewApp()
	app.Name = "crawl-schedule"
	app.HelpName = app.Name
	app.Version = "0.0.1"
	app.ArgsUsage = "<url>"
	app.Usage = "schedules a crawl request in nsq"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "nsq-addr",
			EnvVar: "NSQ_ADDR",
			Usage:  "nsq address (required)",
		},
		cli.StringFlag{
			Name:   "topic",
			EnvVar: "TOPIC",
			Usage:  "crawl requests nsq topic (required)",
			Value:  "crawl_requests",
		},
		cli.StringSliceFlag{
			Name:  "form-value",
			Usage: "form value in format (format: key=value)",
		},
		cli.StringSliceFlag{
			Name:  "metadata",
			Usage: "metadata value in format (format: key=value)",
		},
		cli.StringSliceFlag{
			Name:  "callback",
			Usage: "crawl request callbacks (required)",
		},
		cli.StringFlag{
			Name:  "referer",
			Usage: "crawl request referer",
		},
		cli.StringFlag{
			Name:  "method",
			Usage: "crawl request referer",
			Value: "GET",
		},
	}
	app.Before = func(c *cli.Context) error {
		var errs []string
		if len(c.String("topic")) == 0 {
			errs = append(errs, "Topic cannot be empty")
		}
		if len(c.String("nsq-addr")) == 0 {
			errs = append(errs, "At least one --nsq-addr is required")
		}
		if len(c.StringSlice("callback")) == 0 {
			errs = append(errs, "At least one --callback is required")
		}
		if len(c.Args()) != 1 {
			errs = append(errs, "At least one url is required in arguments.")
		}
		if len(errs) != 0 {
			errs = append([]string{"Errors:"}, errs...)
			return errors.New(strings.Join(errs, "\n"))
		}
		return nil
	}
	app.Action = func(c *cli.Context) {
		form, err := listToMap(c.StringSlice("form-value"))
		if err != nil {
			glog.Fatalf("Form values error: %v", err)
		}

		metadata, err := listToMap(c.StringSlice("metadata"))
		if err != nil {
			glog.Fatalf("Metadata error: %v", err)
		}

		request := &crawl.Request{
			URL:       strings.Trim(c.Args().First(), `"'`),
			Method:    c.String("method"),
			Referer:   c.String("referer"),
			Callbacks: c.StringSlice("callback"),
			Form:      form,
			Metadata:  metadata,
		}

		if glog.V(3) {
			body, _ := json.MarshalIndent(request, "", "  ")
			glog.Infof("Scheduling request: %s", body)
		}

		body, err := json.Marshal(request)
		if err != nil {
			glog.Fatalf("Error marshaling request to json: %v", err)
		}

		cfg := nsq.NewConfig()
		cfg.OutputBufferTimeout = 0

		producer, err := nsq.NewProducer(c.String("nsq-addr"), cfg)
		if err != nil {
			glog.Fatalf("Error connecting to nsq: %v", err)
		}

		producer.SetLogger(log.New(os.Stdout, "[nsq]", 0), nsq.LogLevelError)

		glog.Infof("Publishing request to %q", c.String("topic"))

		if err := producer.Publish(c.String("topic"), body); err != nil {
			glog.Fatalf("Publish error: %v", err)
		}

		glog.Info("Request scheduled")
		producer.Stop()
	}

	if err := app.Run(os.Args); err != nil {
		glog.Fatal(err)
	}
}

func listToMap(list []string) (result map[string]string, err error) {
	result = make(map[string]string)
	for _, keyValue := range list {
		i := strings.Index(keyValue, "=")
		if i <= 0 {
			return nil, fmt.Errorf("%q is not valid", keyValue)
		}
		key := keyValue[:i]
		value := keyValue[i+1:]
		result[key] = value
	}
	return
}
