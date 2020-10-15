# Deploy a testnet‌

This document describes 3 ways to setup a network of `panacead` nodes, each serving a different usecase:

1. Single-node, local
2. Second-node, another machine

## Single-node, local

This guide helps you create a single validator node that runs a network locally for testing and other development related uses.

### Create genesis file and start the network

```bash
# You can run all of these commands from your home directory
cd $HOME

# Initialize the genesis.json file that will help you to bootstrap the network
panacead init testing --chain-id=testing

# Create a key to hold your validator account
panaceacli keys add validator

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom.
panacead add-genesis-account $(panaceacli keys show validator -a) 100000000000000umed

# Generate the transaction that creates your validator
panacead gentx --name validator

# Add the generated bonding transaction to the genesis file
panacead collect-gentxs

# Now its safe to start `panacead`
panacead start
```

This setup puts all the data for `panacead` in `~/.panacead`. You can examine the genesis file you created at `~/.panacead/config/genesis.json`. With this configuration `panaceacli` is also ready to use and has an account with tokens.

## Second-node, another machine

This guide helps you create a second node that connects to the first node which we started before.

### Init use another moniker and same chain id

```bash
panacead init testing2 --chain-id=testing
```

### Overwrite ~/.panacead/config/genesis.json with first node's genesis.json

### Change persistent\_peers

```bash
# If you don't install jq, run "panaceacli status" text and get "id" value manually
panaceacli status | jq .node_info.id 
> 46046c89ec576daa0662613ee0142ab61dd2421e
```

Now paste this id with ip address in `~/.panacead/config/config.toml` .

```text
# Comma separated list of nodes to keep persistent connections to
persistent_peers = "<id>@<first_node_ip>:26656"
```

### Start this second node

```bash
panacead start
```

