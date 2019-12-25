FROM scratch
COPY mc-monitor /
ENTRYPOINT ["/mc-monitor"]