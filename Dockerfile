# syntax=docker/dockerfile:1
FROM golang:1.23 AS builder
WORKDIR /src
COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app .

FROM ubuntu:latest
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
  libssl-dev \
  ca-certificates \
  nftables \
  kmod \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Create startup script to initialize nftables
RUN echo '#!/bin/bash\n\
# Load nftables kernel modules\n\
modprobe nf_tables 2>/dev/null || echo "Warning: Could not load nf_tables module"\n\
modprobe nft_counter 2>/dev/null || echo "Warning: Could not load nft_counter module"\n\
modprobe nft_chain_nat 2>/dev/null || echo "Warning: Could not load nft_chain_nat module"\n\
\n\
# Start the application\n\
exec /ethereum-metrics-exporter "$@"' > /start.sh && chmod +x /start.sh

COPY --from=builder /bin/app /ethereum-metrics-exporter
ENTRYPOINT ["/start.sh"]
