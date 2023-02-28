# Cosmovisor

`cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, `cosmovisor` can automatically download the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

We will explain how to use it based on version `1.4.0` of `cosmovisor`.

## Cosmovisor Setup

Install the `cosmovisor` binary.

You will find a `cosmovisor` in path `$GOPATH/bin`.

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@1.4.0
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
* DAEMON_PREUPGRADE_MAX_RETRIES( Optional)
  * The maximum number of times to call pre-upgrade in the application after exit status of 31. 
  * After the maximum number of retries, cosmovisor fails the upgrade.
  * The default value is `0`

We make setting to manually download the binary and to automatically proceed with the upgrade.

```shell
export DAEMON_HOME=$HOME/.panacea
export DAEMON_NAME=panacead
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
```

Create a directory for the genesis binary and upgrade binary.

```shell
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v2.0.6/bin
```

Clone and check out the current `panacead` mainnet version. (v2.0.5)

```shell
git clone -b v2.0.5 https://github.com/medibloc/panacea-core.git
cd panacea-core
```

Build a binary and copy it to the genesis path of the `cosmovisor`.

```shell
make clean && make build
cp ./build/panacead $DAEMON_HOME/cosmovisor/genesis/bin
```

Build the upgraded version(v2.0.5) and copy it to the `cosmovisor`.

```shell
git checkout tags/v2.0.6
make clean && make build
cp ./build/panacead $DAEMON_HOME/cosmovisor/upgrades/v2.0.6/bin
```

Verify that the `cosmovisor` is applied as the mainnet version normally.

```shell
cosmovisor version
2.0.5 # output
```

Discontinue the currently running `panacead` and run it with `cosmovisor`.

```shell
stop panacead
cosmovisor start
```

When this [proposal](https://www.mintscan.io/medibloc/proposals/10) passes and the upgrade date(`2023-03-08T05:40:00Z`) arrives, the upgrade will automatically proceed to v2.0.6.


## Example

Show an example of upgrading the version of panacead in a local environment.

```shell
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@1.4.0

export DAEMON_HOME=$HOME/.panacea
export DAEMON_NAME=panacead
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false

mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v2.0.6/bin
```

Clone and check out the current `panacead` mainnet version. (v2.0.5)

Then, build it and copy it.
```shell
git clone -b v2.0.5 https://github.com/medibloc/panacea-core.git
cd panacea-core

make clean && make build
cp ./build/panacead $DAEMON_HOME/cosmovisor/genesis/bin
```

Create a new key for the validator, then add a genesis account and transaction:

```shell
./build/panacead init test --chain-id test --overwrite
./build/panacead keys add validator
./build/panacead add-genesis-account validator 1000000000umed --keyring-backend test
./build/panacead gentx validator 1000000umed --chain-id test
./build/panacead collect-gentxs

cat <<< $(jq '.app_state.gov.voting_params.voting_period = "20s" | .app_state.gov.deposit_params.min_deposit[].amount = "10000000"' $DAEMON_HOME/config/genesis.json) > $DAEMON_HOME/config/genesis.json
```

You need to run binary with cosmovisor.
```shell
cosmovisor version
2.0.5 # output
cosmovisor start
```

Submit upgrade proposal along with a deposit and a vote.

```shell
./build/panacead tx gov submit-proposal software-upgrade v2.0.6 \
--title upgrade \
--description upgrade \
--upgrade-height 30 \
--from validator \
--chain-id test \
--yes

./build/panacead tx gov deposit 1 10000000umed --from validator --chain-id test --yes
./build/panacead tx gov vote 1 yes --from validator --chain-id test --yes
```

Build the upgraded version(v2.0.6) and copy it to the `cosmovisor`.

```shell
git checkout tags/v2.0.6
make clean && make build
cp ./build/panacead $DAEMON_HOME/cosmovisor/upgrades/v2.0.6/bin
```

The upgrade will be performed automatically at height 30.

```shell
cosmovisor version
2.0.6 # output
```