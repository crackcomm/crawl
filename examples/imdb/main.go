// This is only an example, please dont harm imdb servers, if you need movies
// data checkout http://www.imdb.com/interfaces I can also recommend checking
// out source code of https://github.com/BurntSushi/goim which implements
// importing data into SQL databases and comes with command line search tool.
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
	flag.Set("logtostderr", "true")
}

var maxReqs = flag.Int("max-reqs", 600, "Max requests per second")

func main() {
	defer log.Flush()
	flag.Parse()

	c := crawl.New(&crawl.Options{
		MaxRequestsPerSecond: *maxReqs,
		QueueCapacity:        100000,
		Verbose:              false,
	})

	spider := new(imdbSpider)
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
	// Entity - Movie entity.
	Entity = "entity"

	// List - Movies list.
	List = "list"
)

type imdbSpider int

func (spider *imdbSpider) List(ctx context.Context, c *crawl.Crawl, resp *crawl.Response) (err error) {
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
func (spider *imdbSpider) Entity(ctx context.Context, c *crawl.Crawl, resp *crawl.Response) (err error) {

	log.Infof(
		"Movie: title=%s year=%s",
		strings.TrimSpace(resp.Query().Find("h1.header span[itemprop=name]:nth-of-type(1)").Text()),
		strings.TrimSpace(resp.Query().Find("h1.header span a").Text()),
	)

	return
}
