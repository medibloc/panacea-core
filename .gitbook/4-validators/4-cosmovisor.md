# Cosmovisor

`cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, `cosmovisor` can automatically download the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

We will explain how to use it based on version `1.4.0` of `cosmovisor`.

For a more detailed explanation, see the [official cosmovisor documentation](https://docs.cosmos.network/main/tooling/cosmovisor).

## Cosmovisor Setup

Install the `cosmovisor` binary.

You will find a `cosmovisor` in path `$GOPATH/bin`.

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest # Current version v1.4.0
```


Set the required environment variabes.

There are four variables. (two are requirement, and two are optional)

* DAEMON_HOME (Requirement)
  * This includes genesis binary and upgrade binaries.
  * `export DAEMON_HOME=$HOME/.panacea`
* DAEMON_NAME (Requirement)
  * It's the name of binary.
  * `export DAEMON_NAME=panacead`
* DAEMON_ALLOW_DOWNLOAD_BINARIES (Optional)
  * Whether to automatically download the new binary.
  * The default value is `false`.
  * Unfortunately, automatic download is not available yet.(by `libwasmvm.so`)
* DAEMON_RESTART_AFTER_UPGRADE (Optional)
  * Whether to restart the process after the upgrade is completed.
  * The default value is `true`.
  * `export DAEMON_RESTART_AFTER_UPGRADE=true`
* DAEMON_RESTART_DELAY (Optional)
  * allow a node operator to define a delay between the node halt (for upgrade) and backup by the specified time. 
  * The value must be a duration.(e.g. 1s)
  * The default value is `none`.
  * `export `DAEMON_RESTART_DELAY=1s`
* DAEMON_DATA_BACKUP_DIR (Optional)
  * Use this option to set a custom backup directory. If not set, $DAEMON_HOME is used.
* UNSAFE_SKIP_BACKUP (Optional)
  * If set to true, upgrades directly without performing a backup. 
  * Otherwise (false, default) backs up the data before trying the upgrade. 
  * The default value of false is useful and recommended in case of failures and when a backup needed to rollback.
  * We recommend using the default backup option UNSAFE_SKIP_BACKUP=false.
* DAEMON_PREUPGRADE_MAX_RETRIES(Optional)
  * The maximum number of times to call pre-upgrade in the application after exit status of 31. 
  * After the maximum number of retries, cosmovisor fails the upgrade.
  * The default value is `0`

We make setting to manually download the binary and to automatically proceed with the upgrade.

## Initialization

Set the necessary environment variables. Below are the default settings for panacea.


```shell
export DAEMON_HOME=$HOME/.panacea
export DAEMON_NAME=panacead
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
```

Create a directory for the genesis binary. This path should contain the panacead currently running on the mainnet.

```shell
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
```

Clone and check out the current `panacead` mainnet version.

```shell
git clone -b tags/[current-active-version] https://github.com/medibloc/panacea-core.git
cd panacea-core
```

Build a binary and copy it to the genesis path of the `cosmovisor`.

```shell
make clean && make build
cp ./build/panacead $DAEMON_HOME/cosmovisor/genesis/bin
```

Discontinue the currently running `panacead` and run it with `cosmovisor`.

```shell
stop panacead
cosmovisor start
```

## Upgrade chain

cosmovisor supports chain upgrades automatically. This requires some preliminary work.

You need to create a path in cosmovisor's home path to store the binary to be upgraded.

```shell
# Generally, [upgrade-version] refers to the 'Name' value set in the SoftwareUpgradeProposal.
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/[upgrade-version]/bin
```

Generate a new binary and navigate to the path you created above.
```
git clone -b tags/[upgrade-version] https://github.com/medibloc/panacea-core.git
cd panacea-core
make clean && make build
cp ./build/panacead $DAEMON_HOME/cosmovisor/upgrades/[upgrade-version]/bin
```

The upgrade is complete when the chain is stopped by SoftwareUpgradeProposal, the binaries are replaced by cosmovisor, and the chain is started.s