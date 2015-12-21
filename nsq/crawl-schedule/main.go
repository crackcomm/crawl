package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
	"github.com/crackcomm/crawl"
	"github.com/golang/glog"
)

func main() {
	app := cli.NewApp()
	app.Name = "crawl-schedule"
	app.HelpName = app.Name
	app.Version = "0.0.1"
	app.Usage = "schedules a crawl request in nsq"
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:   "nsq-addr",
			Usage:  "nsq address (at least one is required)",
			EnvVar: "NSQ_ADDR",
		},
		cli.StringFlag{
			Name:   "requests-topic",
			Value:  "crawl_requests",
			Usage:  "crawl requests topic",
			EnvVar: "REQUESTS_TOPIC",
		},
		cli.StringSliceFlag{
			Name:  "form-value",
			Usage: "form value in format (key=value)",
		},
		cli.StringSliceFlag{
			Name:  "metadata",
			Usage: "metadata value in format (key=value)",
		},
		cli.StringSliceFlag{
			Name:  "callback",
			Usage: "crawl request callbacks (at least one is required)",
		},
		cli.StringFlag{
			Name:  "referer",
			Usage: "crawl request referer",
		},
		cli.StringFlag{
			Name:  "method",
			Value: "GET",
			Usage: "crawl request referer",
		},
	}
	app.Action = func(c *cli.Context) {
		if len(c.StringSlice("nsq-addr")) == 0 {
			glog.Fatalf("At least one --%s is required", "nsq-addr")
		}
		if len(c.StringSlice("callback")) == 0 {
			glog.Fatalf("At least one --callback is required")
		}
		if len(c.Args()) != 1 {
			glog.Fatalf("At least one url is required in arguments.")
		}

		form, err := listToMap(c.StringSlice("form-value"))
		if err != nil {
			glog.Fatalf("Form values error: %v", err)
		}

		metadata, err := listToMap(c.StringSlice("metadata"))
		if err != nil {
			glog.Fatalf("Metadata error: %v", err)
		}

		request := &crawl.Request{
			URL:       strings.Trim(c.Args()[1], `"'`),
			Method:    c.Args()[0],
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

		producer, err := nsq.NewProducer(c.StringSlice("nsq-addr")[0], nsq.NewConfig())
		if err != nil {
			glog.Fatalf("Error connecting to nsq: %v", err)
		}

		if err := producer.Publish(c.String("requests-topic"), body); err != nil {
			glog.Fatalf("Publish error: %v", err)
		}

		glog.Info("Request scheduled")
	}
}

func listToMap(list []string) (result map[string]string, err error) {
	result = make(map[string]string)
	for _, keyValue := range list {
		i := strings.Index(keyValue, "=")
		if i <= 0 {
			return nil, fmt.Errorf("Form value %q is not valid", keyValue)
		}
		key := keyValue[:i-1]
		value := keyValue[i:]
		result[key] = value
	}
	return
}
