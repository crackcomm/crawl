// This is only an example, please dont harm imdb servers, if you need movies
// data checkout http://www.imdb.com/interfaces I can also recommend checking
// out source code of https://github.com/BurntSushi/goim which implements
// importing data into SQL databases and comes with command line search tool.
package main

import (
	"log"

	"golang.org/x/net/context"

	"github.com/crackcomm/crawl"

	imdb "github.com/crackcomm/crawl/examples/imdb/spider"
)

func main() {
	c := crawl.New(
		crawl.WithQueue(crawl.NewQueue(1000)),
		crawl.WithConcurrency(200),
		crawl.WithSpiders(imdb.Spider),
	)

	if err := c.Schedule(context.Background(), &crawl.Request{
		URL:       "http://www.imdb.com/chart/top/",
		Callbacks: crawl.Callbacks(imdb.List),
	}); err != nil {
		log.Fatal(err)
	}

	log.Print("Starting crawl")

	// Its up to You how you want to handle errors
	// You can reschedule request or ignore that
	go func() {
		for err := range c.Errors() {
			log.Printf("Crawl error: %v", err)
		}
	}()

	c.Start()
}
