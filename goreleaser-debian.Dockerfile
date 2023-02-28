FROM debian:latest
COPY ethereum-metrics-exporter* /ethereum-metrics-exporter
ENTRYPOINT ["/ethereum-metrics-exporter"]
