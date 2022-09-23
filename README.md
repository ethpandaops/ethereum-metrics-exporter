# 🦄 Ethereum Address Metrics Exporter 🦄

A Prometheus metrics exporter for Ethereum externally owned account and contract addresses including;

- Externally owned account addresses
- [ERC20](https://eips.ethereum.org/EIPS/eip-20) contracts
- [ERC721](https://eips.ethereum.org/EIPS/eip-20) contracts
- [ERC1155](https://eips.ethereum.org/EIPS/eip-20) contracts
- [Uniswap pair](https://v2.info.uniswap.org/pairs) contracts
- [Chainlink data feed](https://v2.info.uniswap.org/pairs) contracts

# Usage
Ethereum Address Metrics Exporter requires a config file. An example file can be found [here](https://github.com/savid/ethereum-address-metrics-exporter/blob/master/example_config.yaml).

```
A tool to export the ethereum address state

Usage:
  ethereum-address-metrics-exporter [flags]

Flags:
      --config string   config file (default is config.yaml) (default "config.yaml")
  -h, --help            help for ethereum-address-metrics-exporter
```

## Configuration

Ethereum Address Metrics Exporter relies entirely on a single `yaml` config file.

| Name | Default | Description |
| --- | --- | --- |
| global.logging | `warn` | Log level (`panic`, `fatal`, `warn`, `info`, `debug`, `trace`) |
| global.metricsAddr | `:9090` | The address the metrics server will listen on |
| global.namespace | `eth_address` | The prefix added to every metric |
| global.labels[] |  | Key value pair of labels to add to every metric (optional) |
| execution.url | `http://localhost:8545` | URL to the execution node |
| execution.timeout | `10s` | Timeout for requests to the execution node |
| execution.headers[] |  | Key value pair of headers to add on every request |
| addresses.eoa |  | List of ethereum externally owned account addresses |
| addresses.eoa[].name |  | Name of the address, will be a label on the metric |
| addresses.eoa[].address |  | Ethereum externally owned account address |
| addresses.eoa[].labels[] |  | Key value pair of labels to add to this address only (optional) |
| addresses.erc20 |  | List of ethereum [ERC20](https://eips.ethereum.org/EIPS/eip-20) addresses |
| addresses.erc20[].name |  | Name of the address, will be a label on the metric |
| addresses.erc20[].address |  | Ethereum address |
| addresses.erc20[].contract |  | Ethereum contract address |
| addresses.erc20[].labels[] |  | Key value pair of labels to add to this address only (optional) |
| addresses.erc721 |  | List of ethereum [ERC721](https://eips.ethereum.org/EIPS/eip-721) addresses |
| addresses.erc721[].name |  | Name of the address, will be a label on the metric |
| addresses.erc721[].address |  | Ethereum address |
| addresses.erc721[].contract |  | Ethereum contract address |
| addresses.erc721[].labels[] |  | Key value pair of labels to add to this address only (optional) |
| addresses.erc1155 |  | List of ethereum [ERC1155](https://eips.ethereum.org/EIPS/eip-1155) addresses |
| addresses.erc1155[].name |  | Name of the address, will be a label on the metric |
| addresses.erc1155[].address |  | Ethereum address |
| addresses.erc1155[].contract |  | Ethereum contract address |
| addresses.erc1155[].tokenID |  | NFT Token Identifier |
| addresses.erc1155[].labels[] |  | Key value pair of labels to add to this address only (optional) |
| addresses.uniswapPair |  | List of [uniswap pair](https://v2.info.uniswap.org/pairs) addresses |
| addresses.uniswapPair[].name |  | Name of the address, will be a label on the metric |
| addresses.uniswapPair[].from |  | First symbol name, will be a label on the metric |
| addresses.uniswapPair[].to |  | Second symbol name, will be a label on the metric |
| addresses.uniswapPair[].contract |  | Ethereum contract address of the [uniswap pair](https://v2.info.uniswap.org/pairs) |
| addresses.uniswapPair[].labels[] |  | Key value pair of labels to add to this address only (optional) |
| addresses.chainlinkDataFeed |  | List of [chainlink data feed](https://docs.chain.link/docs/ethereum-addresses/) addresses |
| addresses.chainlinkDataFeed[].name |  | Name of the address, will be a label on the metric |
| addresses.chainlinkDataFeed[].from |  | First symbol name, will be a label on the metric |
| addresses.chainlinkDataFeed[].to |  | Second symbol name, will be a label on the metric |
| addresses.chainlinkDataFeed[].contract |  | Ethereum contract address of the [chainlink data feed](https://docs.chain.link/docs/ethereum-addresses/) |
| addresses.chainlinkDataFeed[].labels[] |  | Key value pair of labels to add to this address only (optional) |


### Example

```yaml
global:
  logging: "debug" # panic,fatal,warm,info,debug,trace
  metricsAddr: ":9090"
  namespace: eth_address
  labels:
    extra: label

execution:
  url: "http://localhost:8545"
  timeout: 10s
  headers:
    authorization: "Basic abc123"

addresses:
  eoa:
    - name: John smith
      address: 0x4B1D3c9BEf9D097F564DcD6cdF4558CB389bE3d5
      labels:
        type: friend
    - name: Jane Doe
      address: 0x4B1Df3549940C56d962F248f211788D66B4aAF39
      labels:
        type: acquaintance
        company: NSA
  erc20:
    - name: Some ERC20 Contract
      contract: 0x4B1DB272F63E03Dd37ea45330266AC9328A66DB6
      address: 0x4B1D1465b14cA06e72b942F361Fd3352Aa9c5368
  erc721:
    - name: Some ERC721 Contract
      contract: 0x4B1D23bf5018189fDad68a0E607b6005ccF7E593
      address: 0x4B1DB5c493955C8eF6D2a30CFf47495023b85C8d
  erc1155:
    - name: Some ERC1155 Contract
      contract: 0x4B1D8DC12da8f658FA8BF0cdB18BB7D4dABB2DB3
      tokenID: 100
      address: 0x4B1D6D35f293AB699Bfc6DE141E031F3E3997BBe
  # https://v2.info.uniswap.org/pairs
  uniswapPair:
    - name: eth->usdt
      from: eth
      to: usdt
      contract: 0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852
  # https://docs.chain.link/docs/ethereum-addresses/
  chainlinkDataFeed:
    - name: eth->usd
      from: eth
      to: usd
      contract: 0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419
```

## Getting Started

### Download a release
Download the latest release from the [Releases page](https://github.com/savid/ethereum-address-metrics-exporter/releases). Extract and run with:
```
./ethereum-address-metrics-exporter --config your-config.yaml
```

### Docker
Available as a docker image at [savid/ethereum-address-metrics-exporter](https://hub.docker.com/r/savid/ethereum-address-metrics-exporter/tags)
#### Images
- `latest` - distroless, multiarch
- `latest-debian` - debian, multiarch
- `$version` - distroless, multiarch, pinned to a release (i.e. `0.4.0`)
- `$version-debian` - debian, multiarch, pinned to a release (i.e. `0.4.0-debian`)

**Quick start**
```
docker run -d  --name ethereum-address-metrics-exporter -v $HOST_DIR_CHANGE_ME/config.yaml:/opt/ethereum-address-metrics-exporter/config.yaml -p 9090:9090 -p 5555:5555 -it savid/ethereum-address-metrics-exporter:latest --config /opt/ethereum-address-metrics-exporter/config.yaml;
docker logs -f ethereum-address-metrics-exporter;
```

### Kubernetes via Helm
[Read more](https://github.com/skylenet/ethereum-helm-charts/tree/master/charts/ethereum-address-metrics-exporter)
```
helm repo add ethereum-helm-charts https://skylenet.github.io/ethereum-helm-charts

helm install ethereum-address-metrics-exporter ethereum-helm-charts/ethereum-address-metrics-exporter -f your_values.yaml
```
### Grafana

**Building yourself (requires Go)**

1. Clone the repo
   ```sh
   go get github.com/savid/ethereum-address-metrics-exporter
   ```
2. Change directories
   ```sh
   cd ./ethereum-address-metrics-exporter
   ```
3. Build the binary
   ```sh  
    go build -o ethereum-address-metrics-exporter .
   ```
4. Run the exporter
   ```sh  
    ./ethereum-address-metrics-exporter
   ```

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

### Running locally
#### Backend
```
go run main.go --config your_config.yaml
```
