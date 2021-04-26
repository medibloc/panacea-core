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

Configuration | Value
--------------|-------
Network | Choose an existing VPC or default one if you don't have any
Subnet | Choose an existing subnet or default one if you don't have any
Auto-assign Public IP | Enable only if you access a full node from the outside of its VPC
Other fields | Follow default settings

### Add a storage

Configuration | Value
--------------|-------
Size | 500 GiB
Volume Type | `General Purpose SSD (gp3)`
IOPS | 3000
Throughput | 125 MB/s

### Configure a Security Group

Type | Protocol | Port range |  Description
-----|----------|------------|-------------
SSH | TCP | 22 |
Custom TCP | TCP | 26656 | P2P with other nodes
Custom TCP | TCP | 26657 | RPC
Custom TCP | TCP | 1317 | HTTP

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

timeout_commit = "1s"
```

After that, edit the `~/.panacead/config/app.toml` file in order to enable the anti-spam mechanism and reject incoming transactions with less than a minimum fee:
```toml
# Validators reject any tx from the mempool with less than the minimum fee per gas.
minimum_fees = "0.5umed"

# NOTE: For the Testnet, please set minimum_fees as "", so that no fee is required.
```

Now, your full node has been initialized!

### Copy the Genesis file

Fetch the `genesis.json` file of the latest chain from the following links, and place it to `~/.panacead/config/genesis.json`.
- Mainnet: https://github.com/medibloc/panacea-launch
- Testnet: https://github.com/medibloc/panacea-networks

### Configure Seed Nodes

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

View the status of the network with the Block Explorer
- Mainnet: https://explorer.medibloc.org
- Testnet: https://testnet-explorer.medibloc.org


## Join as a validator

If you want to participate in validating blocks as a validator,
you can register yourself into the validator set by submitting a transaction.

For more details, see the [CLI guide](interaction-with-the-network-cli.md#staking).




