
save-godeps:
	godep save github.com/crackcomm/crawl/nsq/consumer/skeleton

nsq-crawler-build:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o ./dist/crawler main.go

install:
	go install github.com/crackcomm/crawl/nsq/consumer/skeleton

dist: nsq-crawler-build

clean:
	rm -rf dist
