# Initialization

#### Before run the oracle, you should initialize the oracle configuration.

## Preparations

Before initializing the oracle, you should install the oracle described
in [oracle-installation](./1-oracle-installation.md).
After install the oracle, create an empty directory on your host, to be mounted as the `/home_mnt` directory in the
enclave.

```bash
# If you run oracle on your host
sudo mkdir /oracle

# If you run oracle using Docker
mkdir $(pwd)/oracle
```

After that, if you want to run the `oracled` with Docker as described in
the [oracle-installation](./1-oracle-installation.md),
it is recommended to create an environment variable that you can execute the Docker container easily.

```bash
export DOCKER_CMD="docker run --rm \
  --device /dev/sgx_enclave \
  --device /dev/sgx_provision \
  -v $(pwd)/oracle:/oracle ghcr.io/medibloc/panacea-oracle:latest"
```

If you are not going to run the `oracled` without Docker, you can set an `DOCKER_CMD` environment variable empty string.

```bash
export DOCKER_CMD=""
```

## Command Line of Initialization

```bash
$DOCKER_CMD ego run oracled init --home $HOME/.oracle 
```

When run the above CLI for initializing the oracle, the `config.toml` file will be generated under the `$HOME/.oracle`
in the enclave.
The `config.toml` file will be shown like this:

```toml
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Base Configuration                            ###
###############################################################################

log-level = "info"
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

## The Components of `config.toml`

- Base Configuration
    - `log-level`: The log level you want (default: `info` | `debug`)
    - `oracle-mnemonic`: The mnemonic of your oracle account registered in Panacea
    - `oracle-acc-num`: The account number of your oracle account registered in Panacea
    - `oracle-acc-index`: The account index of your oracle account registered in Panacea
    - `data-dir`: The path that you want to store data of light client database
    - `oracle-priv-key-file`: The name of the sealed oracle private key file (default: `oracle_priv_key.sealed`)
    - `oracle-pub-key-file`: The name of the oracle public key json file (default: `oracle_pub_key.json`)
    - `node-priv-key-file`: The name of the sealed oracle node private key file (default: `node_priv_key.sealed`)


- Panacea Configuration
    - `chain-id`: The chain ID of Panacea (mainnet: `panacea-3`)
    - `grpc-addr`: The gRPC address of Panacea (mainnet: `https://grpc.gopanacea.org`)
    - `rpc-addr`: The RPC address of Panacea (maiinet: `https://rpc.gopanacea.org`)
    - `default-gas-limit`: The default gas limit when sending transaction (default: `400000`)
    - `default-fee-amount`: The default fee amount when sending transaction (default: `2000000umed`)
    - `light-client-primary-add`: A primary RPC address for light client verification
    - `light-client-witness-addrs`: Witness addresses (comma-separated) for light client verification
    - `light-client-log-level`: The log information for light client (default: `error` | `warn` | `info`)


- IPFS Configuration
    - `ipfs-node-addr`: Public IPFS node address for store and get data


- API Configuration
    - `listen-addr`: The listen address of the oracle
    - `write-timeout`: The maximum duration before timing out writes of the response (default: `60`)
    - `read-timeout`: The maximum duration for reading the entire request, including the body (default: `15`)


##### After initializing the oracle, you can register oracle based on above configuration.

##### If you want to know how to register oracle, please see the [oracle-registration](./4-oracle-registration.md) documentation.

