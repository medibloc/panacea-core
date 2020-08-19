#!/bin/bash

# A script for deploying two Panacea docker containers on local.
#
# Prerequisites
# - Docker
# - jq

set -euxo pipefail

shopt -s expand_aliases

# Build a docker image, if you don't have
# docker build -t panacea-core .

# Kill docker containers, if exist.
docker kill testing1 testing2 || true

# Create a docker network for Panacea containers.
# They will communicate in this network using their container names as IPs.
docker network rm panacea || true
docker network create panacea

# Create directories which will be mounted to each docker containers.
rm -rf ~/panacea_home
mkdir -p ~/panacea_home/testing1/panacead ~/panacea_home/testing1/panaceacli
mkdir -p ~/panacea_home/testing2/panacead ~/panacea_home/testing2/panaceacli

# A helper function for generating a bash command to run a docker container
function get_cmd() {
  mode_opt=$1
  container_name=$2
  echo "docker run --rm ${mode_opt} --network panacea --name ${container_name} \
  -v $HOME/panacea_home/${container_name}/panacead:/root/.panacead \
  -v $HOME/panacea_home/${container_name}/panaceacli:/root/.panaceacli \
  panacea-core:latest"
}

# Commands for running each Panacea container.
# One is for interactive mode, the other is for detached mode.
cmd1="$(get_cmd "-it" "testing1")"
cmd1_detached="$(get_cmd "-d" "testing1")"
cmd2="$(get_cmd "-it" "testing2")"
cmd2_detached="$(get_cmd "-d" "testing2")"

# Init and run 'testing1' container
$cmd1 panacead init testing1 --chain-id=testing
$cmd1 panaceacli keys add validator
account="$($cmd1 panaceacli keys show validator -a | tr -d '\r')"
$cmd1 panacead add-genesis-account "${account}" 100000000000000umed
$cmd1 panacead gentx --name validator
$cmd1 panacead collect-gentxs
$cmd1_detached panacead start
docker ps
docker logs testing1

# Init and run 'testing2' container
$cmd2 panacead init testing2 --chain-id=testing
cp ~/panacea_home/testing1/panacead/config/genesis.json ~/panacea_home/testing2/panacead/config/
peer_id="$($cmd2 panaceacli status --node tcp://testing1:26657 | jq .node_info.id | sed 's/"//g')"
sed -i '' "s/^persistent_peers[[:space:]]*=.*$/persistent_peers = \"${peer_id}@testing1:26656\"/g" ~/panacea_home/testing2/panacead/config/config.toml
grep "^persistent_peers = " ~/panacea_home/testing2/panacead/config/config.toml
$cmd2_detached panacead start
docker ps
docker logs testing2