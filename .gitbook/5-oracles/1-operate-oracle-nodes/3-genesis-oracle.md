# Genesis Oracle

- Status: Draft
- Created: 2023-01-11
- Modified: 2023-01-11
- Authors
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Youngjoon Lee <yjlee@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Myongsik Gong <myongsik_gong@medibloc.org>
    - Inchul Song <icsong@medibloc.org>
    - Tae Jin Yoon(tj@medibloc.org)


## Synopsis
This document describes how to register a genesis oracle with Panacea.

In order for DEP to operate properly, one or more trusted oracles must be registered in Panacea.
In general, the genesis oracle will already be registered on the testnet or mainnet.
This guide is useful for testing on the localnet.

## Genesis Oracle Registration
To register a genesis oracle, you must first complete the [oracle-initialization](./2-oracle-initialization.md) step.
Panancea needs to complete the steps before executing `panacead start` in [deploy-localnet](../../4-validators/2-deploy-localnet.md).
You must first register the genesis oracle in `genesis.json` of Panacea

### UniqueID
UniqueID can be extracted from Oracle's binary.
Execute the command in oracle as follows:
```shell
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    ghcr.io/medibloc/panacea-oracle:latest \
    ego uniqueid /usr/bin/oracled
```
**Output**
```
EGo v1.0.1 (e1e48c9dbfdfd3cb2f2fda7602729162c9f762cc)
<a-hexadecimal-string-of-unique-id>
```

### Genesis Oracle Registration in Panacea

After the uniqueID extraction is completed, the genesis oracle must be registered in `genesis.json` of Panacea.

We provide a CLI for this process.
```
panacead add-genesis-oracle \
  --oracle-unique-id <unique-id>
  --oracle-account <oracle-address-or-key> \
  --oracle-commission-rate <oracle-commission-rate> \
  --oracle-commission-max-rate <oracle-commission-max-rate> \
  --oracle-commission-max-change-rate <oracle-commission-max-change-rate> \
  --oracle-endpoint {oracle_endpoint}
```

| Argument                          | Requirement | Description                                                                                          |
|-----------------------------------|-------------|------------------------------------------------------------------------------------------------------|
| oracle-unique-id                  | optional    | The uniqueID to be set in the params of the oracle module and the genesis oracle                     |
| oracle-account                    | optional    | The address or key name of the account to be registered as an genesis oracle                         |
| oracle-commission-rate            | optional    | The desired initial oracle commission rate                                                           |
| oracle-commission-max-rate        | optional    | The maximum oracle commission rate. The oracle commission rate cannot be greater than this max rate. |
| oracle-commission-max-change-rate | optional    | The maximum rate that an oracle can change once. It will be reset 24 hours after the last change.    |
| oracle-endpoint                   | optional    | The endpoint of oracle to be used                                                                    |

You can check the oracle registered in `genesis.json`
```
cat $HOME/.panacea/config/genesis.json | jq .app_state.oracle.oracles
```
**Output**
```json
[
  {
    "oracle_address": "<oracle-address>",
    "unique_id": "<unique-id>",
    "endpoint": "<endpoint>",
    "update_time": "<update-time>",
    "oracle_commission_rate": "<commission-rate>",
    "oracle_commission_max_rate": "<commission-max-rate>",
    "oracle_commission_max_change_rate": "<commission-max-change-rate>"
  }
]
```

### Start block generation in Panacea
You need to launch Panacea to start generating blocks. 
```shell
panacead start
```

## Oracle Key Pair and Remote Report

The genesis oracle must create an oracle private key and public key to use for data encryption/decryption.
This oracle also issues a remote report to allow others to prove that the genesis oracle is running inside secure enclave and the oracle key pair is generated inside the enclave.

### Generates oracle key pair and remote report in oracle

The genesis oracle needs to generate an oracle key pair and a remote report.
However, before generating the oracle keys and remote reports, the genesis oracle needs to know the trusted block information from Panacea 
even though this information is not a necessary component of the oracle key pair generation or the remote report generation.
This information is not a necessary component because the two generation processes do not retrieve data from Panacea.

When the oracle participates in the verification operation (`oracled start`), the oracle needs to use a light client as it will retrieve data from Panacea. Therefore, unless trusted block information is received during the process of generating an oracle key, the genesis oracle has no way to retrieve this block information; this is why we retrieve the trusted block information during the oracle key pair and remote report generation.

You can get trusted block information by:
```shell
BLOCK=$(panacead q block --node <node-rpc-address>)

HEIGHT=$(echo $BLOCK | jq -r .block.header.height)
HASH=$(echo $BLOCK | jq -r .block_id.hash)
```

After getting the height and hash of the block, you can generate the necessary keys and remote report via the CLI below.
```
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v <directory-you-want>:/oracle \
    ghcr.io/medibloc/panacea-oracle:latest \
    ego run /usr/bin/oracled gen-oracle-key \
      --trusted-block-height $HEIGHT \
      --trusted-block-hash $HASH
      
```

| Argument             | Requirement | Description                                                 |
|----------------------|-------------|-------------------------------------------------------------|
| trusted-block-height | required    | Trusted block height of Panacea                             |
| trusted-block-hash   | required    | Block hash corresponding to trusted block height of Panacea |


When the oracle key and remote report are generated successfully, they are stored as file with the below structure:

```
# Oracle home
.
├── config.toml
├── data
│   └── light-client.db
│       ├── 000001.log
│       ├── CURRENT
│       ├── LOCK
│       ├── LOG
│       └── MANIFEST-000000
├── oracle_priv_key.sealed
└── oracle_pub_key.json
```
- `data/light-client.db`: Repository of Light client.
- `oracle_priv_key.sealed`: Oracle private key sealed file.
- `oracle_pub_key.json`: A json file containing oracle's public key and remote report.

**oracle_pub_key.json**
```json
{
  "public_key_base64": "<oracle-public-key-base64>",
  "remote_report_base64": "<oracle-remote-report-base64>"
}
```

### Submit a parameter change proposal
The generated oracle public key and its remote report should be set by governance, a proposal for changing module parameter of oracle module.

```shell
panacead tx gov submit-proposal param-change proposal.json \
  --from <key> \
  --chain-id <chain-id> \
  --gas auto \
  --gas-adjustment 1.3 \
  --gas-prices 5umed \
  -y
```
**proposal.json**
```json
{
  "title": "<title>",
  "description": "<description>",
  "changes": [
    {
      "subspace": "oracle",
      "key": "OraclePublicKey",
      "value": "<oracle-pub-key-base64>"
    },
    {
      "subspace": "oracle",
      "key": "OraclePubKeyRemoteReport",
      "value": "<oracle-remote-report-base64>"
    }
  ],
  "deposit": "100000000000umed"
}
```

After submitting the proposal, vote on the proposal and wait for it to pass.
```shell
panacead tx gov vote {proposal_id} yes \
  --from <key> \
  --chain-id <chain_id> \
  --fees 1000000umed \
  -y
```

If the proposal passes, you can check the changes with the following CLI.
```shell
panacead q oracle params
```
**Output**
```shell
params:
  oracle_pub_key_remote_report: "<oracle-remote-report-base64>"
  oracle_public_key: "<oracle-public-key-base64>"
  unique_id: "<unique-id>"
```

When all these processes are completed, the genesis oracle can operate normally.

Next, see [Running node](./5-running-node.md) where you can run the genesis oracle.
