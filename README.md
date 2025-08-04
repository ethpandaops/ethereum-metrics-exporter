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

#### Web3 Metrics
- **`eth_exe_web3_client_version`** - Client version (labels: `version`)

#### Network Metrics
- **`eth_exe_net_peer_count`** - Number of connected peers

### Consensus Layer Metrics (`eth_con_*`)

#### Beacon Block Metrics
- **`eth_con_beacon_slot`** - The slot number in the block
- **`eth_con_beacon_transactions`** - The amount of transactions in the block
- **`eth_con_beacon_slashings`** - The amount of slashings in the block
- **`eth_con_beacon_attestations`** - The amount of attestations in the block
- **`eth_con_beacon_deposits`** - The amount of deposits in the block
- **`eth_con_beacon_voluntary_exits`** - The amount of voluntary exits in the block
- **`eth_con_beacon_finality_checkpoint_epochs`** - The epochs of the finality checkpoints
- **`eth_con_beacon_reorg_count`** - The count of reorgs
- **`eth_con_beacon_reorg_depth`** - The depth of reorgs
- **`eth_con_beacon_proposer_delay`** - The delay of the proposer (histogram)
- **`eth_con_beacon_empty_slots_count`** - The number of slots that have expired without a block proposed
- **`eth_con_beacon_withdrawals`** - The amount of withdrawals in the block
- **`eth_con_beacon_withdrawals_amount_gwei`** - The sum amount of all withdrawals in the block (in gwei)
- **`eth_con_beacon_blob_kzg_commitments`** - The amount of blob kzg commitments in the block

#### Event Metrics
- **`eth_con_event_time_since_last_subscription_event_ms`** - The amount of time since the last subscription event (in milliseconds)

#### General Node Metrics
- **`eth_con_node_version`** - The version of the running beacon node
- **`eth_con_peers`** - The count of peers connected to beacon node

#### Consensus Spec Metrics
- **`eth_con_spec_safe_slots_to_update_justified`** - The number of slots to wait before updating the justified checkpoint
- **`eth_con_spec_deposit_chain_id`** - The chain ID of the deposit contract
- **`eth_con_spec_config_name`** - The name of the config
- **`eth_con_spec_max_validators_per_committee`** - The maximum number of validators per committee
- **`eth_con_spec_seconds_per_eth1_block`** - The number of seconds per ETH1 block
- **`eth_con_spec_base_reward_factor`** - The base reward factor
- **`eth_con_spec_epochs_per_sync_committee_period`** - The number of epochs per sync committee period
- **`eth_con_spec_effective_balance_increment`** - The effective balance increment
- **`eth_con_spec_max_attestations`** - The maximum number of attestations
- **`eth_con_spec_min_sync_committee_participants`** - The minimum number of sync committee participants
- **`eth_con_spec_genesis_delay`** - The number of epochs to wait before processing the genesis block
- **`eth_con_spec_seconds_per_slot`** - The number of seconds per slot
- **`eth_con_spec_max_effective_balance`** - The maximum effective balance
- **`eth_con_spec_terminal_total_difficulty`** - The terminal total difficulty
- **`eth_con_spec_terminal_total_difficulty_trillions`** - The terminal total difficulty in trillions
- **`eth_con_spec_max_deposits`** - The maximum number of deposits
- **`eth_con_spec_min_genesis_active_validator_count`** - The minimum number of active validators at genesis
- **`eth_con_spec_target_committee_size`** - The target committee size
- **`eth_con_spec_sync_committee_size`** - The sync committee size
- **`eth_con_spec_eth1_follow_distance`** - The number of blocks to follow behind the head of the eth1 chain
- **`eth_con_spec_terminal_block_hash_activation_epoch`** - The epoch at which the terminal block hash is activated
- **`eth_con_spec_min_deposit_amount`** - The minimum deposit amount
- **`eth_con_spec_slots_per_epoch`** - The number of slots per epoch
- **`eth_con_spec_preset_base`** - The base of the preset

#### Sync Status Metrics
- **`eth_con_sync_percentage`** - How synced the node is with the network (0-100%)
- **`eth_con_sync_estimated_highest_slot`** - The estimated highest slot of the network
- **`eth_con_sync_head_slot`** - The current slot of the node
- **`eth_con_sync_distance`** - The sync distance of the node
- **`eth_con_sync_is_syncing`** - 1 if the node is in syncing state

#### Health Metrics
- **`eth_con_health_check_results_total`** - Total of health checks results
- **`eth_con_health_up`** - Whether the node is up or not

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
