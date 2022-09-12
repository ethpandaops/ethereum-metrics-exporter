FROM gcr.io/distroless/static-debian11:latest
COPY ethereum-address-metrics-exporter* /ethereum-address-metrics-exporter
ENTRYPOINT ["/ethereum-address-metrics-exporter"]
