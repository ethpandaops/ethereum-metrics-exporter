FROM gcr.io/distroless/static-debian11:latest
COPY ethereum-metrics-exporter* /ethereum-metrics-exporter
ENTRYPOINT ["/ethereum-metrics-exporter"]
