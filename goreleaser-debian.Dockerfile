FROM debian:latest
COPY ethereum-address-metrics-exporter* /ethereum-address-metrics-exporter
ENTRYPOINT ["/ethereum-address-metrics-exporter"]
