# Genesis oracle

- Status: Draft
- Created: 2023-01-11
- Modified: 2023-01-11
- Authors
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Youngjoon Lee <yjlee@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Myongsik Gong <myongsik_gong@medibloc.org>
    - Inchul Song <icsong@medibloc.org>


## Synopsis
This document describes how to register a genesis oracle with Panacea.

In order for DEP to operate properly, one or more trusted oracles must be registered in Panacea.
In general, the genesis oracle will already be registered on the testnet or mainnet.
This guide is useful for testing on the localnet.

## Genesis oracle registration
To register a genesis oracle, you must first complete the [oracle-inialzation](./2-oracle-intialization.md) step.
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
{A hexadecimal string of unique_id}
```

### Genesis oracle registration in Panacea

After uniqueID extraction is completed, the genesis oracle must be registered in `genesis.json` of Panacea.

We provide a CLI for this process.
```
panacead add-genesis-oracle \
  --oracle-unique-id {unique_id} 
  --oracle-account {address or key_name} \
  --oracle-commission-rate {oracle_commission_rate} \
  --oracle-commission-max-rate {oracle_commission_max_rate} \
  --oracle-commission-max-change-rate {oracle_commission_max_change_rate} \
  --oracle-endpoint {oracle_endpoint}
```

| Argument                          | Requirement | Description                                                                                          |
|-----------------------------------|-------------|------------------------------------------------------------------------------------------------------|
| oracle-unique-id                  | false       | The uniqueID to be set in the params of the oracle module and the genesis oracle                     |
| oracle-account                    | false       | The address or key name of the account to be registered as an genesis oracle                         |
| oracle-commission-rate            | false       | The desired initial oracle commission rate                                                           |
| oracle-commission-max-rate        | false       | The maximum oracle commission rate. The oracle commission rate cannot be greater than this max rate. |
| oracle-commission-max-change-rate | false       | The maximum rate that an oracle can change once. It will be reset 24 hours after the last change.    |
| oracle-endpoint                   | false       | The endpoint of oracle to be used                                                                    |

You can check the oracle registered in `genesis.json`
```
cat $HOME/.panacea/config/genesis.json | jq .app_state.oracle.oracles
```
**Output**
```json
[
  {
    "oracle_address": "{oracle_address}",
    "unique_id": "{unique_id}",
    "endpoint": "{endpoint}",
    "update_time": "0001-01-01T00:00:00Z",
    "oracle_commission_rate": "{commission_rate}",
    "oracle_commission_max_rate": "{commission_max_rate}",
    "oracle_commission_max_change_rate": "{commission_max_change_rate}"
  }
]
```

### Start block generation in Panacea
You need to launch the Panacea to start generating blocks. 
```shell
panacead start
```

## Registering oracle keys and remote report

The genesis oracle must create an oracle private key and public key to use for data encryption/decryption.
It also issues a remote report to prove that it is a valid oracle.

### Generates oracle's private key, public key and remote report in oracle

You can get trusted block information by:
```shell
BLOCK=$(panacead q block --node <node-rpc-address>)

HEIGHT=$(echo $BLOCK | jq -r .block.header.height)
HASH=$(echo $BLOCK | jq -r .block_id.hash)
```

With the above arguments, you can generate the necessary keys and remote report via the CLI below.
```
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v {ANY_DIR_ON_HOST}:/oracle \
    ghcr.io/medibloc/panacea-oracle:latest \
    ego run /usr/bin/oracled gen-oracle-key \
      --trusted_block_height $HEIGHT \
      --trusted-block-hash $HASH
      
```

| Argument             | Requirement | Description                                                 |
|----------------------|-------------|-------------------------------------------------------------|
| trusted-block-height | true        | Trusted block height of Panacea                             |
| trusted-block-hash   | true        | Block hash corresponding to trusted block height of Panacea |



When the Oracle key and remote report generation is completed, the file is created with the following structure.

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
  "public_key_base64": "{oracle_public_key_base64}",
  "remote_report_base64": "{oracle_remote_report_base64}"
}
```

### Submit a parameter change proposal
In order to register Oracle's public key and remote report with Panacea, a parameter change proposal must be submit.

```shell
panacead tx gov submit-proposal param-change proposal.json \
  --from {key} \
  --chain-id {chain_id} \
  --fees 1000000umed \
  -y
```
**proposal.json**
```json
{
  "title": "{title}",
  "description": "{description}",
  "changes": [
    {
      "subspace": "oracle",
      "key": "OraclePublicKey",
      "value": "{oracle_pub_key_base64}"
    },
    {
      "subspace": "oracle",
      "key": "OraclePubKeyRemoteReport",
      "value": "{oracle_remote_report_base64}"
    }
  ],
  "deposit": "100000000000umed"
}
```

After submitting the proposal, vote on the proposal and wait for it to pass.
```shell
panacead tx gov vote {proposal_id} yes \
  --from {key} \
  --chain-id {chain_id} \
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
  oracle_pub_key_remote_report: "{oracle_remote_report_base64"
  oracle_public_key: "{oracle_public_key_base64}"
  unique_id: "{unique_id}"
```

When all these processes are completed, Panacea and Oracle can operate DEP normally.