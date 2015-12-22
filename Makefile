
install-crawl-schedule:
	go install github.com/crackcomm/crawl/nsq/crawl-schedule

install: install-crawl-schedule

docs-deps: install
	go install github.com/crackcomm/tdc
	go install github.com/crackcomm/crawl/nsq/consumer/skeleton

docs: docs-deps
	sh -c 'TDC_CRAWL_SCHEDULE_HELP=`crawl-schedule --help` \
		TDC_SKELETON_HELP=`skeleton --help` \
			tdc --input docs-templates/ --output .'
