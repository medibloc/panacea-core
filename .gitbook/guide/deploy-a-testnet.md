# Deploy a testnet

This document describes 2 ways to setup a network of `panacead` nodes.

1. [Using multiple machines](#using-multiple-machines)
2. [Using a single machine (using Docker)](#using-a-single-machine-using-docker)


## Using multiple machines

### Deploy the first node

```bash
# Initialize the genesis.json file that will help you to bootstrap the network
panacead init node1 --chain-id=testing

# Create a key to hold your validator account
panaceacli keys add validator

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom.
panacead add-genesis-account $(panaceacli keys show validator -a) 100000000000000umed

# Generate the transaction that creates your validator
panacead gentx --name validator --amount 1000000000000umed --commission-rate 0.1 --commission-max-rate 0.2 --commission-max-change-rate 0.01  --min-self-delegation 1000000

# Add the generated bonding transaction to the genesis file
panacead collect-gentxs

# Now its safe to start `panacead`
panacead start
```
This setup puts all the data for `panacead` in `~/.panacead`.
You can examine the genesis file that you created at `~/.panacead/config/genesis.json`.
With this configuration, `panaceacli` is also ready to use and has an account with tokens.

### Deploy the second node

Init the second node using another moniker: `node2`.
```bash
panacead init node2 --chain-id=testing
```

Overwrite `~/.panacead/config/genesis.json` with the first node's `genesis.json`.

Get a node ID of the first node.
```bash
panacead tendermint show-node-id
> 46046c89ec576daa0662613ee0142ab61dd2421e
```

Set a `persistent_peers` in the `~/.panacead/config/config.toml` of the second node.
```toml
# Comma separated list of nodes to keep persistent connections to
persistent_peers = "<first_node_id>@<first_node_ip>:26656"
```

Start the second node.
```bash
panacead start
```


## Using a single machine (using Docker)

Multiple nodes can be run in a single machine using Docker.
```bash
# NOTE: Make sure that you already installed jq, docker and docker-compose

docker build -t panacea-core .

./scripts/prepare_docker_compose.sh
docker-compose up
```
6 nodes are started. 4 of them are validators.
You can find data directories from `$HOME/panancea_home/<node_id>/*` in your local which are mounted to each Docker container.

You can get a shell to a container by the following command:
```bash
docker exec -it panacea-core_node1_1 /bin/sh
```
