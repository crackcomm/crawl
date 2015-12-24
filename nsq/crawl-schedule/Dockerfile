FROM busybox
MAINTAINER Łukasz Kurowski <crackcomm@gmail.com>
COPY crawl-schedule /crawl-schedule
ENV NSQ_ADDR nsq.service.local:4150
ENTRYPOINT ["/crawl-schedule"]
