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
   --nsq-addr 							nsq address (required) [$NSQ_ADDR]
   --topic "crawl_requests"					crawl requests nsq topic (required) [$TOPIC]
   --form-value [--form-value option --form-value option]	form value in format (format: key=value)
   --metadata [--metadata option --metadata option]		metadata value in format (format: key=value)
   --callback [--callback option --callback option]		crawl request callbacks (required)
   --referer 							crawl request referer
   --method "GET"						crawl request referer
   --timeout "0"						request timeout
   --help, -h							show help
   --version, -v						print the version
   
```
