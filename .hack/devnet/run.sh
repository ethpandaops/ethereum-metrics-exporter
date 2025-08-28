#!/bin/bash
set -x
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

## Get data volumes and their host paths
BEACON_DATA_VOLUME=$(docker inspect "$FIRST_BEACON_NODE" --format='{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}}{{"\n"}}{{end}}{{end}}' | grep '^data-' | head -n1)
EXECUTION_DATA_VOLUME=$(docker inspect "$FIRST_EXECUTION_NODE" --format='{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}}{{"\n"}}{{end}}{{end}}' | grep '^data-' | head -n1)

BEACON_DATA_HOST_PATH=""
EXECUTION_DATA_HOST_PATH=""

if [ -n "$BEACON_DATA_VOLUME" ]; then
  BEACON_DATA_HOST_PATH=$(docker volume inspect "$BEACON_DATA_VOLUME" --format='{{.Mountpoint}}')
  echo "Beacon node data volume: $BEACON_DATA_VOLUME -> $BEACON_DATA_HOST_PATH"

  # Check if directory is accessible, try OrbStack path if not
  if [ ! -d "$BEACON_DATA_HOST_PATH" ] && [[ "$BEACON_DATA_HOST_PATH" == /var/lib/docker/* ]]; then
    ORBSTACK_BEACON_PATH="${HOME}/OrbStack/docker/volumes/${BEACON_DATA_VOLUME}"
    if [ -d "$ORBSTACK_BEACON_PATH" ]; then
      BEACON_DATA_HOST_PATH="$ORBSTACK_BEACON_PATH"
      echo "  -> Using OrbStack path: $BEACON_DATA_HOST_PATH"
    else
      echo "  -> Warning: Directory not accessible at either location"
    fi
  elif [ ! -d "$BEACON_DATA_HOST_PATH" ]; then
    echo "  -> Warning: Directory not accessible: $BEACON_DATA_HOST_PATH"
  fi
fi

if [ -n "$EXECUTION_DATA_VOLUME" ]; then
  EXECUTION_DATA_HOST_PATH=$(docker volume inspect "$EXECUTION_DATA_VOLUME" --format='{{.Mountpoint}}')
  echo "Execution node data volume: $EXECUTION_DATA_VOLUME -> $EXECUTION_DATA_HOST_PATH"

  # Check if directory is accessible, try OrbStack path if not
  if [ ! -d "$EXECUTION_DATA_HOST_PATH" ] && [[ "$EXECUTION_DATA_HOST_PATH" == /var/lib/docker/* ]]; then
    ORBSTACK_EXECUTION_PATH="${HOME}/OrbStack/docker/volumes/${EXECUTION_DATA_VOLUME}"
    if [ -d "$ORBSTACK_EXECUTION_PATH" ]; then
      EXECUTION_DATA_HOST_PATH="$ORBSTACK_EXECUTION_PATH"
      echo "  -> Using OrbStack path: $EXECUTION_DATA_HOST_PATH"
    else
      echo "  -> Warning: Directory not accessible at either location"
    fi
  elif [ ! -d "$EXECUTION_DATA_HOST_PATH" ]; then
    echo "  -> Warning: Directory not accessible: $EXECUTION_DATA_HOST_PATH"
  fi
fi

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
  enabled: $([ -n "$BEACON_DATA_HOST_PATH" ] || [ -n "$EXECUTION_DATA_HOST_PATH" ] && echo "true" || echo "false")
  interval: 1m  # Polling interval (in minutes) - accepts time units: s, m, h
  directories:$(
    [ -n "$BEACON_DATA_HOST_PATH" ] && echo "
  - $BEACON_DATA_HOST_PATH"
    [ -n "$EXECUTION_DATA_HOST_PATH" ] && echo "
  - $EXECUTION_DATA_HOST_PATH"
    [ -z "$BEACON_DATA_HOST_PATH" ] && [ -z "$EXECUTION_DATA_HOST_PATH" ] && echo "
  - /data/ethereum"
  )
EOF

cat <<EOF
============================================================================================================
ethereum-metrics-exporter config at ${__dir}/generated-ethereum-metrics-exporter-config.yaml
============================================================================================================
EOF
