#!/bin/bash

# A script for initializing multiple Panacea nodes, so that they can be started via Docker.
# This script should work for both Linux and macOS.
#
# Prerequisites
# - Docker
# - jq

set -euxo pipefail
shopt -s expand_aliases

# Configurations
NUM_NODES=6
NUM_VALIDATORS=4
CHAIN_ID=panacea-2
HOME_ROOT=$HOME/panacea_home

# Recreate data directories
rm -rf "${HOME_ROOT}"
for (( i=1; i<=$NUM_NODES; i++ )); do
  mkdir -p "${HOME_ROOT}/${i}/panacead" "${HOME_ROOT}/${i}/panaceacli"
done

# A helper function for generating a bash command to run a docker container
function get_cmd() {
  mode_opt=$1
  node_id=$2
  echo "docker run --rm ${mode_opt} \
  -v ${HOME_ROOT}/${node_id}/panacead:/root/.panacead \
  -v ${HOME_ROOT}/${node_id}/panaceacli:/root/.panaceacli \
  panacea-core:latest"
}

# Init the 1st node and create genesis accounts and transactions (to create validators)
cmd="$(get_cmd "-it" "1")"
$cmd panacead init node1 --chain-id=${CHAIN_ID}
for (( i=1; i<=$NUM_VALIDATORS; i++ )); do
  validator="validator${i}"
  $cmd panaceacli keys add ${validator}
  account="$($cmd panaceacli keys show ${validator} -a | tr -d '\r')"
  $cmd panacead add-genesis-account "${account}" 100000000000000umed
  $cmd mkdir -p /root/.panacead/config/gentx/

  cmdval="$(get_cmd "-it" "${i}")"
  pubkey="$($cmdval panacead tendermint show-validator | tr -d '\r')"
  $cmd panacead gentx --name ${validator} --pubkey ${pubkey} --amount 10000000000000umed --commission-rate 0.1 --commission-max-rate 0.2 --commission-max-change-rate 0.01  --min-self-delegation 1000000 --output-document "/root/.panacead/config/gentx/gentx-${i}.json"
done
$cmd panacead collect-gentxs --gentx-dir /root/.panacead/config/gentx/

# Modify some params in the genesis.json
genesis_path="${HOME_ROOT}/1/panacead/config/genesis.json"
jq '.app_state.staking.params.max_validators = 5' "${genesis_path}" > /tmp/genesis.json && mv /tmp/genesis.json "${genesis_path}"

# Copy the genesis.json to other nodes' data directories
for (( i=2; i<=$NUM_NODES; i++ )); do
  cmd="$(get_cmd "-it" "${i}")"
  $cmd panacead init "node${i}" --chain-id=${CHAIN_ID}
  cp "${HOME_ROOT}/1/panacead/config/genesis.json" "${HOME_ROOT}/${i}/panacead/config/genesis.json"
done

# Assemble a persistent_peers parameter
PERSISTENT_PEERS=""
for (( i=1; i<=$NUM_VALIDATORS; i++ )); do
  cmd="$(get_cmd "-it" "${i}")"
  peer_id="$($cmd panacead tendermint show-node-id | tr -d '\r')"
  if [ ${i} -gt 1 ]; then
    PERSISTENT_PEERS+=","
  fi
  PERSISTENT_PEERS+="${peer_id}@node${i}:26656"
done

# Modify config.toml of all nodes
for (( i=1; i<=$NUM_NODES; i++ )); do
  config_path="${HOME_ROOT}/${i}/panacead/config/config.toml"
  sed -i '' "s|^persistent_peers[[:space:]]*=.*$|persistent_peers = \"${PERSISTENT_PEERS}\"|g" "${config_path}"
  sed -i '' 's|^timeout_commit = .*$|timeout_commit = \"1s\"|g' "${config_path}"
  sed -i '' 's|^size = .*$|size = 5000|g' "${config_path}"
  sed -i '' 's|^pex = .*$|pex = true|g' "${config_path}"

#  config_path="${HOME_ROOT}/${i}/panacead/config/panacead.toml"
#  sed -i '' 's|^minimum-gas-prices = .*$|minimum-gas-prices = \"5.0umed\"|g' "${config_path}"
done
