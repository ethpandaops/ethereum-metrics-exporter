#!/bin/bash
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -f "${__dir}/custom-kurtosis.devnet.config.yaml" ]; then
  config_file="${__dir}/custom-kurtosis.devnet.config.yaml"
else
  config_file="${__dir}/kurtosis.devnet.config.yaml"
fi

## Run devnet using kurtosis
ENCLAVE_NAME="${ENCLAVE_NAME:-ethereum-metrics-exporter}"
ETHEREUM_PACKAGE="${ETHEREUM_PACKAGE:-github.com/ethpandaops/ethereum-package}"
if kurtosis enclave inspect "$ENCLAVE_NAME" > /dev/null; then
  echo "Kurtosis enclave '$ENCLAVE_NAME' is already up."
else
  kurtosis run "$ETHEREUM_PACKAGE" \
  --image-download always \
  --enclave "$ENCLAVE_NAME" \
  --args-file "${config_file}"
fi

## Generate ethereum-metrics-exporter config
ENCLAVE_UUID=$(kurtosis enclave inspect "$ENCLAVE_NAME" --full-uuids | grep 'UUID:' | awk '{print $2}')

FIRST_BEACON_NODE=$(docker ps -aq -f "label=kurtosis_enclave_uuid=$ENCLAVE_UUID" \
              -f "label=com.kurtosistech.app-id=kurtosis" \
              -f "label=com.kurtosistech.custom.ethereum-package.client-type=beacon" | tac | head -n1)

FIRST_EXECUTION_NODE=$(docker ps -aq -f "label=kurtosis_enclave_uuid=$ENCLAVE_UUID" \
              -f "label=com.kurtosistech.app-id=kurtosis" \
              -f "label=com.kurtosistech.custom.ethereum-package.client-type=execution" | tac | head -n1)

cat <<EOF > "${__dir}/generated-ethereum-metrics-exporter-config.yaml"
consensus:
  enabled: true
  url: "$(
    port=$(docker inspect --format='{{ (index (index .NetworkSettings.Ports "4000/tcp") 0).HostPort }}' $FIRST_BEACON_NODE)
    if [ -z "$port" ]; then
      port=$(docker inspect --format='{{ (index (index .NetworkSettings.Ports "3500/tcp") 0).HostPort }}' $FIRST_BEACON_NODE)
    fi
    echo "http://127.0.0.1:$port"
  )"
  name: "consensus-client"
execution:
  enabled: true
  url: "$(
    port=$(docker inspect --format='{{ (index (index .NetworkSettings.Ports "8545/tcp") 0).HostPort }}' $FIRST_EXECUTION_NODE)
    if [ -z "$port" ]; then
      port="65535"
    fi
    echo "http://127.0.0.1:$port"
  )"
  name: "execution-client"
  modules:
    - "eth"
    - "net"
    - "web3"
    - "txpool"
diskUsage:
  enabled: false
  interval: 60m  # Polling interval (in minutes) - accepts time units: s, m, h
  directories:
  - /data/ethereum
EOF

cat <<EOF
============================================================================================================
ethereum-metrics-exporter config at ${__dir}/generated-ethereum-metrics-exporter-config.yaml
============================================================================================================
EOF
