# Incentivized Testnet

MediBloc will open a Incentivized Testnet program for a month.
This program is for validators to be onboarded to the process how to set up nodes and participate in block provisions.


## Timeline

*To be determined soon*

- 10th May ~ 11th June: Operate the Testnet
- Mid-June: Upgrade the Mainnet + Validator Joining


## Environment

A new Testnet will be prepared by MediBloc until May 6th, 2021.
The peer addresses will be announced soon.

The staking parameters will be set as below (can be changed).

|Parameter|Value|
|---------|-----|
|max_validators|50|
|inflation_min|7%|
|inflation_max|10%|
|inflation_rate_change|3%|
|goal_bonded|67%|
|block_per_year|31557600|


## Incentivisation Plan

- `100,000` MED will be paid by MediBloc to all validators when they join the Incentivized Testnet program.
- All validators set up their nodes and participate in block provisions by staking all MEDs that they received.
- After the Testnet program is finished, each validator would have `F` MED by accumulating the following rewards:
  - Block rewards
  - Transaction fees
  - For details, please see the [Validator FAQ](https://hub.cosmos.network/main/validators/validator-faq.html#what-is-the-incentive-to-stake).
- Each validator receives the incentive on the Mainnet in proportion to the rewards they earned on the Testnet.
  - If a validator earned the rewards `R = F - 100000` MED on the Testnet, the validator will receive the incentive on the Mainnet as the following formula:
    ```
    incentive_k = total_pie * (R_k / (R_1 + R_2 + ... + R_n))
    ```
    - `incentive_k`: The incentive that the validator `k` will receive on the Mainnet
    - `R_k`: The rewards that the validator `k` earned on the Testnet
	- `n`: The number of validators whose `R` is `> 0`
  - The `total_pie` is `1,000,000` MED (can be changed).
- If all validators operate their node without any downtime and malicious behaviors, all of them will get the same amount of incentives.
- If a validator misses some block provisions or behave maliciously (such as double signing), they cannot earn full amount of block rewards minted during the Testnet period. Or, they can be slashed. Thus, their incentives on the Mainnet will be decreased. For details, please see the [Validator FAQ](https://hub.cosmos.network/main/validators/validator-faq.html#what-are-the-slashing-conditions).
- Validators whose `R` is `<= 0` will not get any incentive on the Mainnet.


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
