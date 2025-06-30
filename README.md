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
   go get github.com/ethpandaops/ethereum-metrics-exporter
   ```
2. Change directories
   ```sh
   cd ./ethereum-metrics-exporter
   ```
3. Build the binary
   ```sh
    go build -o ethereum-metrics-exporter .
   ```
4. Run the exporter
   ```sh
    ./ethereum-metrics-exporter
   ```

## Available Metrics

<details>
<summary>Click to expand the complete list of available metrics</summary>

### Execution Layer Metrics (`eth_exe_*`)

#### General Metrics
- **`eth_exe_gas_price_gwei`** - Current gas price in gwei
- **`eth_exe_network_id`** - Network ID of the node
- **`eth_exe_chain_id`** - Chain ID of the node

#### Sync Status Metrics
- **`eth_exe_sync_percentage`** - Node sync percentage (0-100%)
- **`eth_exe_sync_starting_block`** - Starting block of sync procedure
- **`eth_exe_sync_current_block`** - Current block of sync procedure
- **`eth_exe_sync_is_syncing`** - 1 if node is syncing
- **`eth_exe_sync_highest_block`** - Highest block of sync procedure

#### Block Metrics
- **`eth_exe_block_most_recent_number`** - Most recent block number (labels: `identifier`)
- **`eth_exe_block_head_gas_used`** - Gas used in most recent block
- **`eth_exe_block_head_gas_limit`** - Gas limit of most recent block
- **`eth_exe_block_head_base_fee_per_gas`** - Base fee per gas in most recent block
- **`eth_exe_block_head_block_size_bytes`** - Size of most recent block in bytes
- **`eth_exe_block_head_transactions_in_block`** - Transactions in most recent block
- **`eth_exe_block_safe_gas_used`** - Gas used in most recent safe block
- **`eth_exe_block_safe_gas_limit`** - Gas limit in most recent safe block
- **`eth_exe_block_safe_base_fee_per_gas`** - Base fee per gas in most recent safe block
- **`eth_exe_block_safe_block_size_bytes`** - Size of most recent safe block in bytes
- **`eth_exe_block_safe_transaction_count`** - Transactions in most recent safe block

#### Transaction Pool Metrics
- **`eth_exe_txpool_transactions`** - Transaction count in txpool (labels: `status` - values: "pending", "queued")

#### Admin Metrics
- **`eth_exe_admin_node_info`** - Node information (labels: `ip`, `listenAddr`, `name`, `discovery_port`, `listener_port`, `network`)
- **`eth_exe_admin_node_port`** - Node ports (labels: `name`, `port_name`)
- **`eth_exe_admin_peers`** - Number of connected peers

#### Web3 Metrics
- **`eth_exe_web3_client_version`** - Client version (labels: `version`)

#### Network Metrics
- **`eth_exe_net_peer_count`** - Number of connected peers

### Consensus Layer Metrics (`eth_con_*`)
*Provided by the integrated beacon client package*
- All standard beacon chain metrics including:
  - Beacon chain sync status
  - Validator performance metrics
  - Attestation metrics
  - Block proposal metrics
  - Network participation metrics
  - Peer count and networking metrics

### Disk Usage Metrics (`eth_disk_*`)
- **`eth_disk_usage_bytes`** - Directory disk usage in bytes (labels: `directory`)

### Constant Labels
**Execution metrics** include: `ethereum_role="execution"`, `node_name={configured}`, `module={module_name}`  
**Consensus metrics** include beacon client standard labels  
**Disk metrics** include: `directory={monitored_directory}`

### Required Modules
Each execution metric group requires specific Ethereum client API modules:
- **General**: `["eth", "net"]`
- **Sync**: `["eth"]`
- **Block**: `["eth", "net"]`
- **TxPool**: `["txpool"]`
- **Admin**: `["admin"]`
- **Web3**: `["web3"]`
- **Net**: `["net"]`

</details>

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
