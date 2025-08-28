# 🦄 Ethereum Metrics Exporter 🦄

> A Prometheus metrics exporter for Ethereum execution & consensus nodes

Ethereum client implementations expose extensive Prometheus metrics however there is minimal standardization around the metrics structure. This makes observability across multiple clients a painful experience. This exporter hopes to help alleviate this problem by creating a client-agnostic set of metrics that operators can run without any additional configuration to dashboards or alerting.

To provide these client-agnostic metrics the exporter relies entirely on these well-defined APIs:
- Execution clients
  - [JSON-RPC](https://geth.ethereum.org/docs/rpc/server)
- Consensus clients
  - [Beacon Node API](https://ethereum.github.io/beacon-APIs/#/)

Naturally this means that the exporter is limited to metrics that are exposed by these APIs.

## Built With

* [pf13/cobra-cli](https://github.com/spf13/cobra-cli)
* [ethereum/go-ethereum](https://github.com/ethereum/go-ethereum)
* [attestantio/go-eth2-client](github.com/attestantio/go-eth2-client)
## Usage

```
A tool to export the state of ethereum nodes

Usage:
  ethereum-metrics-exporter [flags]

Flags:
      --config string                   config file (default is $HOME/.ethereum-metrics-exporter.yaml)
      --consensus-url string            (optional) URL to the consensus node
      --disk-usage-interval string      (optional) interval for disk usage metrics collection (e.g. 1h, 5m, 30s)
      --execution-modules strings       (optional) execution modules that are enabled on the node
      --execution-url string            (optional) URL to the execution node
  -h, --help                            help for ethereum-metrics-exporter
      --metrics-port int                Port to serve Prometheus metrics on (default 9090)
      --monitored-directories strings   (optional) directories to monitor for disk usage
  -t, --toggle                          Help message for toggle
```
## Getting Started

### Grafana
* [Single instance dashboard](https://grafana.com/grafana/dashboards/16277)

### Docker
Available as a docker image at `ethpandaops/ethereum-metrics-exporter`

#### Images
- `latest` - distroless, multiarch
- `debian-latest` - debian, multiarch
- `$version` - distroless, multiarch, pinned to a release (i.e. 0.4.0)
- `$version-debian` - debian, multiarch, pinned to a release (i.e. 0.4.0-debian)

**Quick start**
```
docker run -d -it --name ethereum-metrics-exporter -p 9090:9090 -it ethpandaops/ethereum-metrics-exporter --consensus-url=http://localhost:5052 --execution-url=http://localhost:8545
````
**With a config file**
```
docker run -d -it --name ethereum-metrics-exporter -v $HOST_DIR_CHANGE_ME/config.yaml:/opt/exporter/config.yaml -p 9090:9090 -it ethpandaops/ethereum-metrics-exporter --config /opt/exporter/config.yaml

```
### Kubernetes via Helm
[Read more](https://github.com/skylenet/ethereum-helm-charts/tree/master/charts/ethereum-metrics-exporter)
```
helm repo add ethereum-helm-charts https://ethpandaops.github.io/ethereum-helm-charts

helm install ethereum-metrics-exporter ethereum-helm-charts/ethereum-metrics-exporter -f your_values.yaml
```

### Standalone
**Downloading a release**
Available [here](https://github.com/ethpandaops/ethereum-metrics-exporter/releases)

**Building yourself (requires Go)**

1. Clone the repo
   ```sh
   git clone https://github.com/ethpandaops/ethereum-metrics-exporter.git
   ```
2. Change directories
   ```sh
   cd ./ethereum-metrics-exporter
   ```
3. Build the binary
   ```sh
   make build
   ```
4. Run the exporter
   ```sh
   ./build/ethereum-metrics-exporter
   ```

## Development

This project includes a Makefile to simplify common development tasks. Here are the available commands:

### Running the full stack

Note: To run the full stack, you need to have [kurtosis installed](https://docs.kurtosis.com/install).

```sh
# This starts up a kurtosis devnet with a beacon node and an execution nodes
# It will then also start the exporter and expose the metrics of the first beacon and execution node on port 9090
make devnet-run
```

### Building

```sh
# Build for current platform
make build

# Build for Linux amd64
make build-linux

# Build for all platforms (using goreleaser)
make build-all

# Install to $GOPATH/bin
make install
```

### Testing & Quality

```sh
# Run tests with race detection and coverage
make test

# Run short tests only
make test-short

# Generate coverage report
make coverage

# Run linting
make lint

# Run linting with auto-fix
make lint-fix

# Run go vet
make vet

# Format code
make fmt
```

### Dependencies

```sh
# Download dependencies
make deps

# Tidy dependencies
make tidy

# Verify dependencies
make verify
```

### Releases

```sh
# Create a release (requires proper git tags and GitHub token)
make release

# Create a snapshot release for testing
make release-snapshot
```

### Docker

```sh
# Build Docker image
make docker-build

# Push Docker image
make docker-push
```

### Other Commands

```sh
# Run the application directly
make run

# Clean build artifacts
make clean

# Check if required tools are installed
make check-tools

# Display help for all commands
make help
```

### Prerequisites

Before developing, ensure you have the following tools installed:

- Go 1.22+ (check with `go version`)
- golangci-lint (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- goreleaser (see [installation guide](https://goreleaser.com/install/))
- Docker (for building container images)

### Screenshots
![Example](./example.png)
## Contributing

Contributions are greatly appreciated! Pull requests will be reviewed and merged promptly if you're interested in improving the exporter!

1. Fork the project
2. Create your feature branch:
    - `git checkout -b feat/new-metric-profit`
3. Commit your changes:
    - `git commit -m 'feat(profit): Export new metric: profit`
4. Push to the branch:
    -`git push origin feat/new-metric-profit`
5. Open a pull request

## Contact

Sam - [@samcmau](https://twitter.com/samcmau)
