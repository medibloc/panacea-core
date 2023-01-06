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

When data providers [submit consent to provide their data](./3-data-provider-consent.md), the budget deposited by the data consumer is distributed to the provider and the oracle that verified the data.

### Motivation

This is for transparent distribution for the cost of providing data and the cost of verifying data.

### Definitions

`Data Provider`, `Data Consumer` and `Oracle` are defined in [User Flow](./1-user-flow.md)

## Technical Specification

### Reward Distribution

When [creating a data deal](./2-data-deal.md#create-data-deal), data consumers specify the quantity and the budget for the data.
From this, the price per data can be calculated as:

```
price_per_data = deposit / max_num_data
```

Of the `price_per_data`, as an oracle commission for data verification, the commission rate set by the oracle is transferred to the oracle, and the rest is transferred to the data provider.

```
oracle_reward = price_per_data * oracle_commission_rate
provider_reward = price_per_data * (1 - oracle_commission_rate)
```

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