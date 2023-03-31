# Oracle Upgrade

This document is about the entire process of oracle upgrade.

## Oracle Upgrade Process

All oracles registered in Panacea are forced to run the oracle with the same unique ID stored in the oracle module.
Since the new version (to be upgraded) of oracle will have a unique ID different from the existing one, a new unique ID should be registered in Panacea.
This process is determined by on-chain governance; thus, an upgrade can be proposed by an `oracle-upgrade` proposal.

### Submit Proposal for Oracle Upgrade

Oracle upgrade can be proposed by submitting an `oracle-upgrade` proposal as shown below:

```bash
panacead tx gov submit-proposal oracle-upgrade \
  --title "<proposal-title>" \
  --description  "<proposal-description>" \
  --upgrade-height <upgrade-target-height> \
  --upgrade-unique-id "<oracle-unique-id>" \
  --deposit "<deposit-in-umed>" \
  --from "<proposer-key-name>" \
  --chain-id "<chain-id>" \ 
  --fees "1000000umed"
```

- title: The title of oracle upgrade proposal
- description: A description of oracle upgrade proposal
- upgrade-height: A target height to be upgraded
- upgrade-unique-id: The unique ID of oracle to be upgraded
- deposit: A deposit for proposal

If the proposal passes, you can check the oracle upgrade information with the below CLI.

```bash
panacead q oracle oracle-upgrade-info
```

### Upgrade Oracle Node

{% hint style="info" %}
You can upgrade your oracle any time after an `oracle-upgrade` proposal has passed (even before the upgrade target height is reached).
However, since the new version of oracle can be active after the target height, the current version of oracle must be running before the target height.

You do not have to stop the old version of oracle when upgrading to a new version of oracle.
You can run different versions of oracle at the same time using the `home` flag.
{% endhint %}

#### Initialization

The `oracle_priv_key.sealed` used in the previous version cannot be used in the new version of oracle because it cannot be decrypted in the new version of oracle. For more info, refer to [this](../../3-protocol-devs/1-dep-specs/5-confidential-oracle.md)

Therefore, the new version of oracle should also know the oracle private key.
The oracle can retrieve the oracle private key with a similar process to registering an oracle.

Let's start with [initialization](2-oracle-intialization.md) of the new version of oracle.

```bash
export ORACLE_CMD="docker run --rm \
  --device /dev/sgx_enclave \
  --device /dev/sgx_provision \
  -v <directory-you-want>/oracle:/oracle ghcr.io/medibloc/panacea-oracle:main \
  ego run /usr/bin/oracled"
  
  $ORACLE_CMD init --home /home_mnt/.oracle-new
```

You can rename the path where you want to store the configuration file of the new version of oracle.
For this document, we will use `.oracle-new`.
After initialization, complete the configuration by referring to the [Configuring Some Default Setting](2-oracle-intialization#configuring-some-default-setting.md).

#### Request to Upgrade Oracle

The purpose of this request is to securely receive the oracle private key for a new version of oracle.
It is similar to sharing the oracle private key in [oracle registration](4-oracle-registration#request-to-register-oracle.md).
Prior to the request, trusted block information is also required.
You can get trusted block information with the following command:

```bash
BLOCK=$(panacead q block --node <node-rpc-address>)

HEIGHT=$(echo $BLOCK | jq -r .block.header.height)
HASH=$(echo $BLOCK | jq -r .block_id.hash)
```

Then, request an upgrade using this trusted block information.

```bash
$ORACLE_CMD upgrade-oracle \ 
    --trusted-block-height ${HEIGHT} \
    --trusted-block-hash ${HASH}
    --home /home_mnt/.oracle-new
```

Like oracle registration, a `Node Key` to be used for sharing the oracle private key will be generated and stored.
You will find that the `node_priv_key.sealed` file is stored in `<directory-you-want>/oracle/.oracle-new`.
This `node_priv_key.sealed` file is also necessary to retrieve the oracle private key, so it is highly recommended to manage it safely.

#### Subscribe Approval of Upgrade

If the transaction for `upgrade-oracle` succeeds, the oracle will start subscribing to the `ApproveOracleUpgradeEvent`.
Upon approval by other oracles, this new version of oracle can retrieve the oracle private key.
This process is similar to [oracle-registration](4-oracle-registration#subscribe-approval-of-registration.md), so please refer it for details.

#### Running the Upgraded Oracle

After reaching the target height of the upgrade, you can start the new version of oracle.

```bash
$ORACLE_CMD start --home /home_mnt/.oracle-new
```

See [running-oracle](5-running-node.md) for more information.
