# Incentivized Testnet

MediBloc will open a Incentivized Testnet program for a month.
This program is for validators to be onboarded to the process how to set up nodes and participate in block provisions.


## Environment

A Testnet is currently being operated with 3 validator nodes by MediBloc.
It will be upgraded with the following parameters for the Incentivized Testnet.

|Parameter|Value|
|---------|-----|
|max_validators|50|
|inflation_min|7%|
|inflation_max|10%|
|inflation_rate_change|3%|
|goal_bonded|67%|
|block_per_year|31557600|

For parameters that are not listed above, please see a [genesis.json](https://github.com/medibloc/panacea-networks/tree/master/hygieia-4) of the latest Testnet chain.


## Incentivisation Plan

- The `i` MEDs will be paid by MediBloc to all validators when they join the Incentivized Testnet program.
  - **The `i` will be determined soon by MediBloc.**
- All validators set up their nodes and participate in block provisions by staking all MEDs that they received.
- After the Testnet program is finished, each validator would have `f` MEDs by accumulating the following rewards:
  - Block rewards
  - Transaction fees
  - For details, please see the [Validator FAQ](https://hub.cosmos.network/main/validators/validator-faq.html#what-is-the-incentive-to-stake).
- The incentive that the validator will be paid is `f - i` MEDs. The incentive will be paid when the validator join the Mainnet.
- If all validators operate their node without any downtime and malicious behaviors, all of them will get the same amount of incentives.
  - If a validator misses some block provisions or behave maliciously (such as double signing),
    - they cannot earn full amount of block rewards minted during the Testnet period.
    - they can be slashed.
    - For details, please see the [Validator FAQ](https://hub.cosmos.network/main/validators/validator-faq.html#what-are-the-slashing-conditions)
- If their `f` is smaller than `i`, they will not get any incentive.


## Guide for Validators

Please see the [Join the Network](../guide/join-the-network.md) guide.


## Transaction Simulation

MediBloc will generate following transactions with dummy data for the realistic testing.
mediBloc will set 1 MED as a transaction fee.

- DID
  - Creating/Updating/Deactivating DIDs
- AOL (Append-Only Log)
  - Create topics
  - Adding writers
  - Adding records
- Fund
  - Transferring funds between accounts created by MediBloc
- Delegation
  - Delegating/Undelegating to 3 validators operated by MediBloc.
