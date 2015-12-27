# crawl-schedule

Command line tool for scheduling crawl requests in nsq.

## Usage

```sh
$ go install github.com/crackcomm/crawl/nsq/crawl-schedule
$ # or
$ make install-crawl-schedule
$ crawl-schedule --help
NAME:
   crawl-schedule - schedules a crawl request in nsq

USAGE:
   crawl-schedule [global options] command [command options] <url>
   
VERSION:
   0.0.1
   
COMMANDS:
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --nsq-addr 							nsq address [$NSQ_ADDR]
   --topic "crawl_requests"					crawl requests nsq topic [$TOPIC]
   --form-value [--form-value option --form-value option]	form value in format (key=value)
   --metadata [--metadata option --metadata option]		metadata value in format (key=value)
   --callback [--callback option --callback option]		crawl request callbacks (at least one is required)
   --referer 							crawl request referer
   --method "GET"						crawl request referer
   --help, -h							show help
   --version, -v						print the version
   
```
