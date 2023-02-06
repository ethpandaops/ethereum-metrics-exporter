FROM debian:latest
COPY ethereum-metrics-exporter* /exporter
ENTRYPOINT ["/exporter"]
