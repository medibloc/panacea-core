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
  - Tae Jin Yoon <tj@medibloc.org>


## Synopsis

When a data deal is created, a budget for the data is deposited from the data consumer account. 
The deposit, proposed by the data consumer, is distributed to the data providers and the oracles that verified the data when the data provider submits a consent with the certificate to Panacea.  
Data consumers can [deactivate the deal](./2-data-deal.md#Deactivate-Data-Deal) that they created and retrieve the remaining budget at any time if the deal is in status `DEAL_STATUS_ACTIVE`.

### Motivation

This is for a transparent distribution of the incentives for providing the data and verifying the data.

### Definitions

`Data Provider`, `Data Consumer` and `Oracle` are defined in [User Flow](./1-user-flow.md)

## Technical Specification

### Budget Deposit

When data consumers [create a deal](./2-data-deal.md#create-data-deal), they specify a budget for data provision.
At this time, the total amount of the budget is transferred from the data consumer account to the deal account.

### Reward Distribution

When creating a data deal, data consumers specify the quantity as well as the budget for the data.
From these two values, the price per data can be calculated as:

```
price_per_data = deposit / max_num_data
```

When a data provider submits a consent and a certificate, a portion of the `price_per_data`, is transferred to the oracle and the rest is transferred to the data provider. 
This portion is associated with the oracle commision rate that can be set differently for each oracle. 

The data consumer can find out which oracle verified the data by referring to the [certificate](./4-data-validation.md#Response-Body) submitted by the data provider and how much commission fee was paid.

```
oracle_reward = price_per_data * oracle_commission_rate
provider_reward = price_per_data * (1 - oracle_commission_rate)
```

### Budget Refund

If consumers want to stop receiving the data and get refund for the rest of their budget, they can deactivate the deal they created.
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
