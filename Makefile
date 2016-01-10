
REPO ?= github.com/crackcomm/crawl
OUTPUT ?= ./dist

build-crawl-schedule:
	mkdir -p $(OUTPUT)/crawl-schedule
	cp ./nsq/crawl-schedule/Dockerfile $(OUTPUT)/crawl-schedule/
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-s -extldflags "-static"' -a -installsuffix cgo \
		-o $(OUTPUT)/crawl-schedule/crawl-schedule ./nsq/crawl-schedule/main.go

dist: clean build-crawl-schedule

clean:
	rm -rf dist

install-crawl-schedule:
	go install $(REPO)/nsq/crawl-schedule

install: install-crawl-schedule

docker-images: dist
	cd $(OUTPUT)/crawl-schedule && docker build -t crackcomm/crawl-schedule .

docker-deploy: docker-images
	docker push crackcomm/crawl-schedule:latest

docs-deps:
	go install github.com/crackcomm/tdc

docs: docs-deps
	sh -c 'TDC_CRAWL_SCHEDULE_HELP=`go run nsq/crawl-schedule/main.go --help` \
		TDC_SKELETON_HELP=`go run nsq/consumer/skeleton/main.go --help` \
			tdc --input docs-templates/ --output .'

example:
	go run examples/imdb/main.go
