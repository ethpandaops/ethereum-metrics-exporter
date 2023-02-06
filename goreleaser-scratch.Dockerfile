FROM gcr.io/distroless/static-debian11:latest
COPY ethereum-metrics-exporter* /exporter
ENTRYPOINT ["/exporter"]
