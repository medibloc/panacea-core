# Join the network

## Setting Up a New Node

These instructions are for setting up a brand new full node from scratch.

First, initialize the node and create the necessary config files:

```bash
panacead init <your_custom_moniker>
```

{% hint style="warning" %}
Note Monikers can contain only ASCII characters. Using Unicode characters will render your node unreachable.
{% endhint %}

Then, modify the `timeout_commit` in the `~/.panacead/config/config.toml` as below.
```toml
[consensus]

timeout_commit = "1s"
```

After that, edit the `~/.panacead/config/app.toml` file in order to enable the anti-spam mechanism and reject incoming transactions with less than a minimum fee:
```toml
# Validators reject any tx from the mempool with less than the minimum fee per gas.
minimum_fees = "0.5umed"
```

Now, your full node has been initialized!

## Genesis & Seeds

### Copy the Genesis File

Fetch the `genesis.json` file of the latest chain from the following links, and place it to `~/.panacead/config/genesis.json`.
- Mainnet: https://github.com/medibloc/panacea-launch
- Testnet: https://github.com/medibloc/panacea-networks

### Add Seed Nodes

Your node needs to know how to find peers.

Seed nodes can be found in:
- Mainnet: https://github.com/medibloc/panacea-launch#persistent-peers
- Testnet: https://github.com/medibloc/panacea-networks#persistent-peers

Insert those `<node_id>@<ip>`s with 26656 port to the `persistent_peers` field in `~/.panacead/config/config.toml`.
```toml
# Comma separated list of nodes to keep persistent connections to
persistent_peers = "8c41cc8a6fc59f05138ae6c16a9eec05d601ef71@13.209.177.91:26656,cc0285c4d9cec8489f8bfed0a749dd8636406a0d@54.180.169.37:26656,1fc4a41660986ee22106445b67444ec094221e76@52.78.132.151:26656"
```

For more information on seeds and peers, see the [Using Tendermint: Peers](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#peers).

## Run a Full Node

Start the full node with this command:

```bash
panacead start
```

Check that everything is running smoothly:

```bash
panaceacli status
```

View the status of the network with the [Block Explorer](https://explorer.medibloc.org).

## Export State

Panacea can dump the entire application state to a JSON file, which could be useful for manual analysis and can also be used as the genesis file of a new network.

Export state with:

```bash
panacead export > [filename].json
```

You can also export state from a particular height \(at the end of processing the block of that height\):

```bash
panacead export --height [height] > [filename].json
```

