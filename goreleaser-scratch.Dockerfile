# NOTE: This distroless image does not support Docker port bandwidth monitoring 
# as it lacks nftables and required system packages. Use goreleaser-debian.Dockerfile 
# or Dockerfile for full functionality.
FROM gcr.io/distroless/static-debian11:latest
COPY ethereum-metrics-exporter* /ethereum-metrics-exporter
ENTRYPOINT ["/ethereum-metrics-exporter"]
