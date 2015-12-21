// Package spider implements imdb spider.
package spider

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/crackcomm/crawl"
	"golang.org/x/net/context"
)

// Movie - IMDB movie.
var Movie = "imdb_movie"

// List - IMDB movies list.
var List = "imdb_list"

// Register - Registers imdb spider.
func Register(c crawl.Crawler) {
	spider := &imdbSpider{Crawler: c}
	c.Register(Movie, spider.Movie)
	c.Register(List, spider.List)
}

type imdbSpider struct {
	crawl.Crawler
}

func (spider *imdbSpider) List(ctx context.Context, resp *crawl.Response) (err error) {
	if err := spider.checkError(ctx, resp); err != nil {
		return err
	}

	resp.Query().Find("table.chart td.titleColumn a").Each(func(_ int, link *goquery.Selection) {
		href, _ := link.Attr("href")
		spider.Crawler.Schedule(ctx, &crawl.Request{
			URL:       href,
			Source:    resp,
			Callbacks: crawl.Callbacks(Movie),
		})
	})

	return
}

func (spider *imdbSpider) Movie(ctx context.Context, resp *crawl.Response) (err error) {
	if err := spider.checkError(ctx, resp); err != nil {
		return err
	}

	title := crawl.Text(resp, "h1.header span[itemprop=name]:nth-of-type(1)")
	year := crawl.Text(resp, "h1.header span a")
	log.Printf("Response: status=%q title=%q year=%s", resp.GetStatus(), title, year)

	return
}

func (spider *imdbSpider) checkError(ctx context.Context, resp *crawl.Response) (err error) {
	doh := crawl.Text(resp, "h1")
	if doh == "D'oh!" {
		return fmt.Errorf("IMDB returned: %q", crawl.Text(resp, "body"))
	}
	return
}
