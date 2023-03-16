# Data Deal

- Status: Draft
- Created: 2023-01-03
- Modified: 2023-03-14
- Authors
  - Youngjoon Lee <yjlee@medibloc.org>
  - Gyuguen Jang <gyuguen.jang@medibloc.org>
  - Hansol Lee <hansol@medibloc.org>
  - Myongsik Gong <myongsik_gong@medibloc.org>
  - Inchul Song <icsong@medibloc.org>
  - Tae Jin Yoon <tj@medibloc.org>


## Synopsis

This document defines a data deal which is a contract for data collecting and payment for data provision in [DEP](../../1-users/3-data-exchange/0-about-dep.md).
Data consumers create data deals by specifying the type, the quantity, and the pricing of the data that they are willing to consume.
Data providers can choose and participate in the deals that match the data that they have when they are willing to provide.
When a data provider provides data to the deal, the payout is distributed to the provider and the oracle that validated the data.
Also, data consumers can deactivate their data deal whenever they want and the remaining budget would be refunded to the consumer's account.

### Motivation

Data consumers want different types of data, and even for the same type of data, they all differ in the desired quantity and desired cost of the data. 
Thus, the data deal was devised so that data consumers can determine the type of data they want, the quantity they want, and the cost level they want.

### Definitions

`Data Provider`, `Data Consumer` and `Oracle` are defined in [User Flow](./1-user-flow.md)

## Technical Specification

Data consumers should be able to post the information described below publicly, so that any data provider can see it. 
Also, data providers should be assured that a particular data consumer really posted the information. 
To meet these requirements, it is recommended to use a public decentralized state machine as a single source of truth, such as Panacea.

### Data Structure of Deal

The structure of data deal.

```proto
message Deal {
  uint64 id = 1;
  string address = 2;
  repeated string data_schema = 3;
  bytes presentation_definition = 4;
  cosmos.base.v1beta1.Coin budget = 5;
  uint64 max_num_data = 6;
  uint64 cur_num_data = 7;
  string consumer_address = 8;
  repeated AgreementTerm agreement_terms = 9;
  DealStatus status = 10;
  string consumer_service_endpoint = 11;
}

message AgreementTerm {
  uint32 id = 1;
  bool required = 2;
  string title = 3;
  string description = 4;
}
```

- `id`: Auto increment id
- `address`: An address of deal generated when deal is created
- `data_schema`: A list of URLs of desired data schema
- `presentation_definition`: Objects that specify the conditions required for verifiable presentation
- `budget`: A budget for consuming data
- `max_num_data`: The maximum number of data the consumer want
- `cur_num_data`: The current number of data provided
- `consumer_address`: A consumer's account address of the form `panacea1...`
- `agreement_terms`: Terms of agreement of data provision
- `status`: The status of deal. 3 statuses can be possible
  - `DEAL_STATUS_ACTIVE`: The status when deal is active (`cur_num_data` < `max_num_data`).  
  - `DEAL_STATUS_INACTIVE`: The status when deal is deactivated (when consumer deactivated the deal)
  - `DEAL_STATUS_COMPLETED`: The status when deal is completed (`max_num_data` of data is provided)
- `consumer_service_endpoint`: The URL of a consumer service that can serve as consumer data storage.

### Create Data Deal

Data consumers can create their deal with the followings:

```proto
message MsgCreateDeal {
  repeated string data_schema = 1;
  cosmos.base.v1beta1.Coin budget = 2;
  uint64 max_num_data = 3;
  string consumer_address = 4;
  repeated AgreementTerm agreement_terms = 5;
  bytes presentation_definition = 6;
  string consumer_service_endpoint =7;
}
```

When deal is created, the amount of budget is sent from consumer's account to deal's account.
In other words, the balance of consumer's account should be greater or equal than the budget.

#### Data Requirements
There are two ways for a data consumer to specify the requirements of the data they want to consume: `data_schema` and `presentation_definition`

**Data Schema**

`data_shcema` is a way to specify a data with [JSON Schema](https://json-schema.org/).

When `data_schema` is set, oracle checks to see if the provider's data satisfies the JSON schema.
Data consumer need to create a json schema, upload it to a specific URI, and put it into the deal.
[Here](http://jsonschema.gopanacea.org/vaccination-cert.json) is the example of json schema.

**Presentation Definition**

`presentation_definition` is a way to specify data in the form of a verifiable presentation.

When `presentation_definition` is set, o ORACLE checks that the provider's data is generated from a VC issued by a Certificate Authority and satisfies the requirements in the presentation definition.

Data consumer can define a `presentation definition` in the following [this](https://identity.foundation/presentation-exchange/#presentation-definition).
Data Consumer need to convert the defined json form of `presentation defintion` to byte array form and put it into the deal.

Below is an example of a `presentation definition`:
```json
{
  "id": "c1b88ce1-8460-4baf-8f16-4759a2f055fd",
  "purpose": "To get data on Korean people aged 18-30",
  "input_descriptors": [
    {
      "id": "age_descriptor",
      "purpose": "Your age should be greater or equal to 18.",
      "constraints": {
        "limit_disclosure": "required",
        "fields": [
          {
            "path": [
              "$.credentialSubject.age"
            ],
            "filter": {
              "type": "integer",
              "minimum": 18,
              "maximum": 30
            }
          },
          {
            "path": [
              "$.credentialSubject.nationality"
            ],
            "filter": {
              "type": "string",
              "enum": [
                "Korea"
              ]
            }
          }
        ]
      }
    }
  ]
}
```

### Deactivate Data Deal

The consumer who created the deal can deactivate the deal at any time as long as `max_num_data` of data is not provided.

To deactivate a deal, the id of the deal should be specified.

```proto
message MsgDeactivateDeal {
  uint64 deal_id = 1;
  string requester_address = 2;
}
```

When a deal is deactivated, all remaining budget is refunded to the data consumer's account.
After the deal is deactivated, data providers cannot provide their data to this deal, and the status of the deal changes to `DEAL_STATUS_INACTIVE`.

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
- 2023-03-14: Add `presentation definition` and `consumer service endpoint` to deal

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).
