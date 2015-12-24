
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

docs-deps: install
	go install github.com/crackcomm/tdc
	go install $(REPO)/nsq/consumer/skeleton

docker-images: dist
	cd $(OUTPUT)/crawl-schedule && docker build -t crackcomm/crawl-schedule .

docker-deploy: docker-images
	docker push crackcomm/crawl-schedule:latest

docs: docs-deps
	sh -c 'TDC_CRAWL_SCHEDULE_HELP=`crawl-schedule --help` \
		TDC_SKELETON_HELP=`skeleton --help` \
			tdc --input docs-templates/ --output .'
