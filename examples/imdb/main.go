package main

import (
	"flag"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/crackcomm/crawl"
	"golang.org/x/net/context"

	log "github.com/golang/glog"
)

func init() {
	flag.Set("v", "2")
	flag.Set("logtostderr", "true")
}

var maxReqs = flag.Int("max-reqs", 600, "Max requests per second")

func main() {
	defer log.Flush()
	flag.Parse()

	c := crawl.New()
	c.SetOptions(&crawl.Options{
		MaxRequestsPerMinute: *maxReqs,
		QueueCapacity:        100000,
	})

	spider := new(Spider)
	c.Handler(Entity, spider.Entity)
	c.Handler(List, spider.List)

	// Add first request
	c.Schedule(&crawl.Request{
		URL:       "http://www.imdb.com/chart/top",
		Callbacks: crawl.Callbacks(List),
	})

	log.Info("Starting crawl")
	if err := c.Start(); err != nil {
		log.Fatal(err)
	}
}

var (
	Entity = "entity"
	List   = "list"
)

type Spider int

func (spider *Spider) List(ctx context.Context, c *crawl.Crawl, resp *crawl.Response) (err error) {
	resp.Query().Find("table.chart td.titleColumn a").Each(func(_ int, link *goquery.Selection) {
		href, _ := link.Attr("href")
		c.Schedule(&crawl.Request{
			URL:       href,
			Source:    resp,
			Callbacks: crawl.Callbacks(Entity),
		})
	})

	return
}

func (spider *Spider) Entity(ctx context.Context, c *crawl.Crawl, resp *crawl.Response) (err error) {

	log.Infof(
		"Movie: %s (%s)",
		strings.TrimSpace(resp.Query().Find("h1.header span[itemprop=name]:nth-of-type(1)").Text()),
		strings.TrimSpace(resp.Query().Find("h1.header span a").Text()),
	)

	return
}
