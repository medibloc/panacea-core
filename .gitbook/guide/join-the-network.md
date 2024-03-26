# Join the Network

This tutorial introduces deploying a new node on [AWS](https://aws.amazon.com/) and join the Panacea Mainnet.


## Launch an AWS EC2 Instance

### Choose an AMI

Choose Ubuntu Server 20.04 LTS 64-bit (x86) with SSD Volume Type.

![](../assets/fullnode/ec2-ami.png)

### Choose the instance type

Choose the `m5.large` instance type (minimum spec).

![](../assets/fullnode/ec2-instance-type.png)

### Configure instance details

| Configuration         | Value |
|-----------------------|------- |
| Network               | Choose an existing VPC or default one if you don't have any |
| Subnet                | Choose an existing subnet or default one if you don't have any |
| Auto-assign Public IP | Enable only if you access a full node from the outside of its VPC |
| Other fields          | Follow default settings |

### Add a storage

| Configuration | Value |
|---------------|------- |
 | Size          | 500 GiB |
 | Volume Type   | `General Purpose SSD (gp3)` |
 | IOPS          | 3000 |
 | Throughput    | 125 MB/s |

### Configure a Security Group

| Type        | Protocol | Port range |  Description |
|-------------|----------|------------|------------- |
| SSH         | TCP | 22 |
| Custom TCP  | TCP | 26656 | P2P with other nodes |
| Custom TCP  | TCP | 26657 | RPC |
| Custom TCP | TCP | 1317 | HTTP |

The P2P `26656` port must be exposed to other Panacea nodes.
If your node will be in the VPC guarded by Sentry nodes, expose `26656` to only Sentry nodes (recommended).
If not, expose it to anywhere.
For details about Sentry nodes, please see the [Tendermint guide](https://docs.tendermint.com/master/nodes/validators.html#local-configuration).

The RPC `26657` and HTTP `1317` ports are for sending transactions/queries to your node.
So, expose them to the network where you perform operational actions.


### Connect to your EC2 instance and install prerequisites.

```bash
ssh ubuntu@<your-ec2-ip> -i <your-key>.pem
```

Install prerequisites by following the [Installation](installation.md) guide.


## Setup a New Node

These instructions are for setting up a brand new full node from scratch.

First, initialize the node and create the necessary config files:

```bash
panacead init <your_custom_moniker>
```

{% hint style="warning" %}
The `moniker` can contains only ASCII characters. Using Unicode characters will render your node unreachable.
{% endhint %}

Then, modify the `timeout_commit` in the `~/.panacead/config/config.toml` as below.
```toml
[consensus]

timeout_commit = "5s"
```

After that, edit the `~/.panacead/config/app.toml` file in order to enable the anti-spam mechanism and reject incoming transactions with less than the `minimum-gas-prices`:
```toml
# Validators reject any tx from the mempool with less than the minimum-gas-prices.
minimum-gas-prices = "5umed"

# NOTE: For the Testnet, please set minimum-gas-prices as "", so that no fee is required.
```

Now, your full node has been initialized!

### Copy the Genesis file

Fetch the `genesis.json` file of the latest chain from the following links, and place it to `~/.panacead/config/genesis.json`.
- Mainnet: https://github.com/medibloc/panacea-mainnet
- Testnet: https://github.com/medibloc/panacea-testnet

### Configure Seed Nodes

Your node needs to know how to find peers.

Seed nodes can be found in:
- Mainnet: https://github.com/medibloc/panacea-mainnet#seed-nodes
- Testnet: https://github.com/medibloc/panacea-testnet#seed-nodes

Insert those `<node_id>@<ip>`s with 26656 port to the `persistent_peers` field in `~/.panacead/config/config.toml`.
```toml
# Comma separated list of nodes to keep persistent connections to
persistent_peers = "8c41cc8a6fc59f05138ae6c16a9eec05d601ef71@13.209.177.91:26656,cc0285c4d9cec8489f8bfed0a749dd8636406a0d@54.180.169.37:26656,1fc4a41660986ee22106445b67444ec094221e76@52.78.132.151:26656"
```

For more information on seeds and peers, see the [Using Tendermint: Peers](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#peers).

### State Sync

Your node can rapidly sync with the network using state sync without replaying historical blocks. For more details, please refer to [this](https://docs.tendermint.com/v0.34/tendermint-core/state-sync.html).

To set state sync enabled, RPC servers and trusted block info (height and hash) are required.

You can use the following public RPC endpoints provided by Medibloc team.
- 3.35.82.40:26657
- 13.124.96.254:26657

trusted block info can be obtained via RPC.

```shell
curl -s 15.165.191.68:26657/block | jq -r '.result.block.header.height + "\n" + .result.block_id.hash'
# 7700000 (height)
# 0D3E53F02ABCDDA8AAC1520342D37A290DDABE4C28190EE6E2C6B0C819F74D4A (hash)
```

Then, you need to edit several things in `~/.panacea/config/config.toml` file.

```toml
[statesync]

enable = true

rpc_servers = "15.165.191.68:26657,54.254.66.59:26657" # rpc addresses
trust_height = <trusted-block-height>
trust_hash = "<trusted-block-hash>"
trust_period = "336h0m0s" # 2/3 of 21 days (unbonding period)
```

If your node have block history data previously synced, you need to clear the data first.

```shell
panacead tendermint unsafe-reset-all
```

## Run a Full Node

Start the full node with this command:

```bash
panacead start
```

Check that everything is running smoothly:

```bash
panacead status
```

View the status of the network with the Block Explorer
- Mainnet: https://explorer.medibloc.org
- Testnet: https://testnet-explorer.medibloc.org

## Join as a validator

If you want to participate in validating blocks as a validator,
you can register yourself into the validator set by submitting a transaction.

For more details, see the [CLI guide](interaction-with-the-network-cli.md#staking).




