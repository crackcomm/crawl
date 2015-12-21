package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/golang/glog"

	"github.com/crackcomm/crawl/nsq/consumer"
	
	imdb "github.com/crackcomm/crawl/examples/imdb/spider"
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

	// Start consumer
	consumer.New(imdb.Register).Run(os.Args)
}
