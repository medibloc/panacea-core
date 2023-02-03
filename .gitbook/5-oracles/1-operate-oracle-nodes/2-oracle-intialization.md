# Oracle Initialization

#### Before running the oracle, you should initialize the oracle configuration.

## Preparations

Before initializing the oracle, you should install the oracle described
in [oracle-installation](./1-oracle-installation.md).
After installing the oracle, create an empty directory on your host, to be mounted as the `/home_mnt` directory in the
enclave.

```bash
mkdir <directory-you-want>/oracle
```

```bash
export ORACLE_CMD="docker run --rm \
  --device /dev/sgx_enclave \
  --device /dev/sgx_provision \
  -v <directory-you-want>/oracle:/oracle ghcr.io/medibloc/panacea-oracle:latest \
  ego run /usr/bin/oracled"
```

## Command Line of Initialization

```bash
$ORACLE_CMD init --home /home_mnt/.oracle 
```

When running the above CLI for initializing the oracle, the `config.toml` file will be generated under
the `<directory-you-want>/.oracle` in the enclave.
The default `config.toml` file will look like below.

```toml
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
# Write comment later
###############################################################################
###                           Base Configuration                            ###
###############################################################################

log-level = "info"
oracle-mnemonic = ""
oracle-acc-num = "0"
oracle-acc-index = "0"
data-dir = "data"

oracle-priv-key-file = "oracle_priv_key.sealed"
oracle-pub-key-file = "oracle_pub_key.json"
node-priv-key-file = "node_priv_key.sealed"

###############################################################################
###                         Panacea Configuration                           ###
###############################################################################

[panacea]

chain-id = ""
grpc-addr = "http://127.0.0.1:9090"
rpc-addr = "tcp://127.0.0.1:26657"
default-gas-limit = "400000"
default-fee-amount = "2000000umed"

# A primary RPC address for light client verification

light-client-primary-addr = "tcp://127.0.0.1:26657"

# Witness addresses (comma-separated) for light client verification

light-client-witness-addrs = "tcp://127.0.0.1:26657"

# Setting log information for light client

light-client-log-level = "error"

###############################################################################
###                         IPFS Configuration                           ###
###############################################################################

[ipfs]

ipfs-node-addr = "127.0.0.1:5001"

###############################################################################
###                         API Configuration                           ###
###############################################################################

[api]

listen-addr = "127.0.0.1:8080"
write-timeout = "60"
read-timeout = "15"
```

## Configuring Some Default Setting

#### Base Configuration

In `Base Configuration`, you need to configure `oracle-mnemonic`, `oracle-acc-num`, and `oracle-acc-index`. These
components should correspond to the account that you registered in Panacea.

#### Panacea Configuration

In `Panacea Configuration`, you need to configure the chain ID of Panacea. If you want to join as an oracle to Panacea
mainnet or testnet, please check the chain IDs
in [mainnet-repoistory](https://github.com/medibloc/panacea-mainnet#intro)
or [testnet-repository](https://github.com/medibloc/panacea-testnet#intro), respectively.

The default `grpc-addr` and `rpc-addr` setting is based on the localnet. So if you want to connect with Panacea
mainnet, the `grpc-addr` should be `https://grpc.gopanacea.org` and the `rpc-addr` should be `https://rpc.gopanacea.org`.
Also, the `light-client-primary-addr` and `light-client-witness-addrs` are as same as `rpc-addr`, if you want to
connect with Panacea mainnet.

The `default-gas-limit` and `default-fee-amount` are set as `400000` and `2000000umed`, since the remote report has
large bytes for oracle registration and oracle upgrade. Once you finish an oracle registration or an upgrade, you
could set a lower gas limit and fee amount than the default setting.

#### IPFS Configuration

The oracle will use a public IPFS node for now. If you want to run a local IPFS node, the `ipfs-node-addr` is same as
default setting. Also, you need to check that the IPFS gateway and the oracle `listen-addr` are at the same port. You can
change the IPFS gateway in `$HOME/.ipfs/config`. If you want to know more about RPC API of the IPFS, please refer
the [IPFS documentation](https://docs.ipfs.tech/reference/kubo/rpc/).

## Next

If you have done the oracle initialization, you could register the oracle based on above configuration. If you want to know
how to register oracle, please refer to the [oracle-registration](./4-oracle-registration.md) documentation.

