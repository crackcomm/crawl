package consumer

import (
	"github.com/codegangsta/cli"
	"github.com/crackcomm/crawl"
)

// Option - Consumer app option setter.
type Option func(*App)

// WithSpiderConstructor - Constructs a spider on action.
func WithSpiderConstructor(fnc func(*App) Spider) Option {
	return func(app *App) {
		app.spiderConstructors = append(app.spiderConstructors, fnc)
	}
}

// WithCrawlerConstructor - Constructs a crawler on action.
func WithCrawlerConstructor(fnc func(*App) crawl.Crawler) Option {
	return func(app *App) {
		app.crawlerConstructor = fnc
	}
}

// WithCrawler - Registers spider on a crawler.
func WithCrawler(crawler crawl.Crawler) Option {
	return func(app *App) {
		app.crawler = crawler
	}
}

// WithSpiders - Registers spider on a crawler.
// It has to be set after WithCrawler (if any).
func WithSpiders(spiders ...Spider) Option {
	return func(app *App) {
		for _, spider := range spiders {
			spider(app.Crawler())
		}
	}
}

// WithBefore - Overwrites flag checking before action.
func WithBefore(fnc func(*cli.Context) error) Option {
	return func(app *App) {
		app.before = fnc
	}
}
