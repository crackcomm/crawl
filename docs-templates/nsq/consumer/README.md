# nsq app

This library is simply a command line application that can be constructed using your own [spiders](https://godoc.org/github.com/crackcomm/crawl/nsq/consumer#Spider).
Example spider can be found in [crawl repo](https://github.com/crackcomm/crawl/blob/master/examples/imdb/spider/spider.go).

### Skeleton

To create your own nsq crawler you can simply copy [`skeleton/`](https://github.com/crackcomm/crawl/tree/master/nsq/consumer/skeleton)
directory and add your own spiders.

After copying it should be enough to replace all occurences of `github.com/crackcomm/crawl/skeleton` to new path of the application.
If you want [CircleCI](https://circleci.com/) to deploy docker image for you change `crawl/skeleton` to your image name in `circle.yaml`.

### Command-line Usage

```sh
$ skeleton --help
{{ .SKELETON_HELP }}
```
