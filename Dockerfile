FROM alpine:3.11
COPY mc-monitor /usr/bin/
ENTRYPOINT ["/usr/bin/mc-monitor"]