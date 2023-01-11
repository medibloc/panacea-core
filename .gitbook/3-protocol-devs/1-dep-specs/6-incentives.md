# Incentives

- Status: Draft
- Created: 2023-01-06
- Modified: 2023-01-06
- Authors
    - Youngjoon Lee <yjlee@medibloc.org>
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Myongsik Gong <myongsik_gong@medibloc.org>
    - Inchul Song <icsong@medibloc.org>


## Synopsis

When data deal is created, a budget for data is deposited from data consumer account. 
And the deposit is distributed to the data providers and the oracles that verified the data.
Data consumers also can [deactivate the deal](./2-data-deal.md#Deactivate-Data-Deal) they created and refund the remaining budget at any time if the deal is in status `DEAL_STATUS_ACTIVE`.

### Motivation

This is for transparent distribution for the cost of providing data and the cost of verifying data.

### Definitions

`Data Provider`, `Data Consumer` and `Oracle` are defined in [User Flow](./1-user-flow.md)

## Technical Specification

### Budget Deposit

When data consumers [create a deal](./2-data-deal.md#create-data-deal), they specify the budget for data provision.
At this time, the amount of the budget is transferred from consumer account to deal account.

### Reward Distribution

When creating a data deal, data consumers specify the quantity as well as the budget for the data.
From this, the price per data can be calculated as:

```
price_per_data = deposit / max_num_data
```

Of the `price_per_data`, as an oracle commission for data verification, the commission rate set by the oracle is transferred to the oracle, and the rest is transferred to the data provider.

The oracle commission can be set differently for each oracle.
you can find out which oracle verified the data by referring to the [certificate](./4-data-validation.md#Response-Body) submitted by the data provider and how much commission fee to be paid.

```
oracle_reward = price_per_data * oracle_commission_rate
provider_reward = price_per_data * (1 - oracle_commission_rate)
```

### Budget Refund

If consumers want to stop being provided data and refund for the rest of their budget, they can deactivate the deal they created.
However, in order to deactivate the deal, the deal must be in the `DEAL_STATUS_ACTIVE` state.

## Backwards Compatibility

Not applicable.

## Forwards Compatibility

Not applicable.

## Example Implementations

Coming soon.

## Other Implementations

None at present.

## History

- 2023-01-06: Initial draft finished

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).