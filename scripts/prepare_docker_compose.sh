#!/bin/bash

set -euxo pipefail

shopt -s expand_aliases

NUM_NODES=6
NUM_VALIDATORS=4
CHAIN_ID=panacea-2
DOCKER_NETWORK=panacea
HOME_ROOT=$HOME/panacea_home

rm -rf "${HOME_ROOT}"
for (( i=1; i<=$NUM_NODES; i++ )); do
  mkdir -p "${HOME_ROOT}/${i}/panacead" "${HOME_ROOT}/${i}/panaceacli"
done

# Create a docker network for Panacea containers.
# They will communicate in this network using their container names as IPs.
docker network rm ${DOCKER_NETWORK} || true
docker network create ${DOCKER_NETWORK}

# A helper function for generating a bash command to run a docker container
function get_cmd() {
  mode_opt=$1
  node_id=$2
  echo "docker run --rm ${mode_opt} --network ${DOCKER_NETWORK} --name node${node_id} \
  -v ${HOME_ROOT}/${node_id}/panacead:/root/.panacead \
  -v ${HOME_ROOT}/${node_id}/panaceacli:/root/.panaceacli \
  panacea-core:latest"
}

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

for (( i=2; i<=$NUM_NODES; i++ )); do
  cmd="$(get_cmd "-it" "${i}")"
  $cmd panacead init "node${i}" --chain-id=${CHAIN_ID}
  cp "${HOME_ROOT}/1/panacead/config/genesis.json" "${HOME_ROOT}/${i}/panacead/config/genesis.json"
done

PERSISTENT_PEERS=""
for (( i=1; i<=$NUM_VALIDATORS; i++ )); do
  cmd="$(get_cmd "-it" "${i}")"
  peer_id="$($cmd panacead tendermint show-node-id | tr -d '\r')"
  if [ ${i} -gt 1 ]; then
    PERSISTENT_PEERS+=","
  fi
  PERSISTENT_PEERS+="${peer_id}@node${i}:26656"
done

for (( i=1; i<=$NUM_NODES; i++ )); do
  cmd="$(get_cmd "-it" "${i}")"
  sed -i '' "s/^persistent_peers[[:space:]]*=.*$/persistent_peers = \"${PERSISTENT_PEERS}\"/g" "${HOME_ROOT}/${i}/panacead/config/config.toml"
done




