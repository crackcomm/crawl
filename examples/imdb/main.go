// This is only an example, please dont harm imdb servers, if you need movies
// data checkout http://www.imdb.com/interfaces I can also recommend checking
// out source code of https://github.com/BurntSushi/goim which implements
// importing data into SQL databases and comes with command line search tool.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/crackcomm/crawl"
	"golang.org/x/net/context"
)

var maxReqs = flag.Int("max-reqs", 600, "Max requests per second")

func main() {
	flag.Parse()

	c := crawl.New(
		crawl.WithQueue(crawl.NewQueue(1000)),
		crawl.WithConcurrency(200),
	)

	spider := &imdbSpider{Crawl: c}

	c.Handler(Entity, spider.Entity)
	c.Handler(List, spider.List)
	c.Schedule(&crawl.Request{
		URL:       "http://www.imdb.com/chart/top",
		Callbacks: crawl.Callbacks(List),
	})

	log.Print("Starting crawl")

	// Its up to You how you want to handle errors
	// You can reschedule request or ignore that
	go func() {
		for err := range c.Errors {
			log.Printf("Error: %v while requesting %q", err.Error, err.Request.String())
		}
	}()

	if err := c.Start(); err != nil {
		log.Fatal(err)
	}
}

var (
	// Entity - Movie entity.
	Entity = "entity"

	// List - Movies list.
	List = "list"
)

type imdbSpider struct {
	*crawl.Crawl
}

func (spider *imdbSpider) List(ctx context.Context, resp *crawl.Response) (err error) {
	if err := spider.checkError(ctx, resp); err != nil {
		return err
	}

	resp.Query().Find("table.chart td.titleColumn a").Each(func(_ int, link *goquery.Selection) {
		href, _ := link.Attr("href")
		spider.Crawl.Schedule(&crawl.Request{
			URL:       href,
			Source:    resp,
			Callbacks: crawl.Callbacks(Entity),
		})
	})

	return
}

func (spider *imdbSpider) Entity(ctx context.Context, resp *crawl.Response) (err error) {
	if err := spider.checkError(ctx, resp); err != nil {
		return err
	}

	title := crawl.Text(resp, "h1.header span[itemprop=name]:nth-of-type(1)")
	year := crawl.Text(resp, "h1.header span a")
	log.Printf("Response: status=%q title=%q year=%s", resp.GetStatus(), title, year)

	return
}

func (spider *imdbSpider) checkError(ctx context.Context, resp *crawl.Response) (err error) {
	doh := crawl.Text("h1")
	if doh == "D'oh!" {
		return fmt.Errorf("IMDB returned: %q", crawl.Text("body"))
	}
	return
}
