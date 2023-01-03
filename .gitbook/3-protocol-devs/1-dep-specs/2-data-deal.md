# Data Deal

- Status: Draft
- Crated: 2023-01-03
- Modified: 2023-01-03
- Authors
  - Youngjoon Lee <yjlee@medibloc.org>
  - Gyuguen Jang <gyuguen.jang@medibloc.org>
  - Hansol Lee <hansol@medibloc.org>
  - Myongsik Gong <myongsik_gong@medibloc.org>
  - Inchul Song <icsong@medibloc.org>


## Synopsis

This document defines a data deal which is a contract for data collecting and pay for provision in [DEP](../../1-users/3-data-exchange/0-about-dep.md).
Data consumers create data deal by specifying what type of data and how much they want, and how much they are willing to pay for it.
Data providers can provide their data to the deals that match the type of data they have.
When a data provider provides data to the deal, the payout is distributed to the provider and oracle that validated the data.
Also, data consumers can deactivate their data deal whenever they want and the remained budget would be refunded to the consumer's account.

### Motivation

Data consumers want different types of data, and even for the same type of data, they all differ in how much they want and how much they are willing to pay.
Thus, the data deal was devised so that data consumers can determine the type of data they want, as much as they want, at a cost they want.

### Definitions

`Data Provider`, `Data Consumer` and `Oracle` are defined in [User Flow](./1-user-flow.md)

## Technical Specification

### Data Structure of Deal

The structure of data deal.

```proto
message Deal {
  uint64 id = 1;
  string address = 2;
  repeated string data_schema = 3;
  cosmos.base.v1beta1.Coin budget = 4;
  uint64 max_num_data = 5;
  uint64 cur_num_data = 6;
  string consumer_address = 7;
  DealStatus status = 8;
}
```

- `id`: auto increment id
- `address`: an address of deal generated when deal is created
- `data_schema`: a list of urls of desired type of data schema
- `budget`: budget for data provision
- `max_num_data`: maximum number of data the consumer want
- `cur_num_data`: current number of data of the deal
- `consumer_address`: consumer's account address (panacea1...)
- `status`: the status of deal. 3 statuses can be possible
  - `DEAL_STATUS_ACTIVE`: the status when deal is active (`cur_num_data` < `max_num_data`).  
  - `DEAL_STATUS_INACTIVE`: the status when deal is deactivated (when consumer deactivated the deal)
  - `DEAL_STATUS_COMPLETED`: the status when deal is completed (`max_num_data` of data is provided)

### Create Data Deal

Data consumers can create their deal with the followings:

```proto
message MsgCreateDeal {
  repeated string data_schema = 1;
  cosmos.base.v1beta1.Coin budget = 2;
  uint64 max_num_data = 3;
  string consumer_address = 4;
}
```

When deal is created, the amount of budget is sent from consumer's account to deal's account.
In other words, the balance of consumer's account should be greater or equal than the budget.

### Deactivate Data Deal

The consumer who created the deal can deactivate the deal at any time as long as `max_num_data` of data is not provided.

To deactivate deal, the id of deal should be specified.

```proto
message MsgDeactivateDeal {
  uint64 deal_id = 1;
  string requester_address = 2;
}
```

When deal is deactivated, all the remained budget is refunded to the consumer's account.
After deal is deactivated, no providers can provide their data to this deal, and the status of the deal would be `DEAL_STATUS_INACTIVE`.

## Backwards Compatibility

Not applicable.

## Forwards Compatibility

For now, JSON schema validation is used for data validation.
It can be expanded using JSON-LD contexts for verifiable credentials in the future.

## Example Implementations

Coming soon.

## Other Implementations

None at present.

## History

- 2023-01-03: Initial draft finished

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).