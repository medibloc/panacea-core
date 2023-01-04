# Data Provider Consent

- Status: Draft
- Crated: 2023-01-04
- Modified: 2023-01-04
- Authors
    - Youngjoon Lee <yjlee@medibloc.org>
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Myongsik Gong <myongsik_gong@medibloc.org>
    - Inchul Song <icsong@medibloc.org>


## Synopsis

This document is about providing data through [DEP](../../1-users/3-data-exchange/0-about-dep.md).
Before providers provide their data to consumers, the data must be validated from oracle whether the data is the correct data specified in the deal.
Providers can send a request for validation to one of the oracles registered in panacea, and the oracle will validate the data using confidential computing without any exposure of the data.
For the data that has been validated, oracle will issue a certificate to the provider.
Providers can consent to provide their data by submitting the certificate to panacea, and will be rewarded by providing the data.
In all of these processes, the data is transmitted with encryption and stored off-chain.

### Motivation

Data should be provided by providers based on their data ownership, and the reward should be distributed in transparent and fair manner.
To do so, providers use digital signature as consent to provide the data.

### Definitions

- `Data Provider`, `Data Consumer`, and `Oracle` are defined in [User Flow](./1-user-flow.md)
- `Deal` is defined in [Data Deal](./2-data-deal.md)
- `Certificate`: a certificate that the data is validated to be provided to the deal, which is issued by oracle.

## Technical Specification

Before provider provides the data to deal, data should be validated by the oracle.
If the data is successfully validated to be provided, the provider will have a `Certificate` like below (more about [Data Validation](./4-data-validation.md)):

```proto
message Certificate {
  UnsignedCertificate unsigned_certificate = 1;
  bytes signature = 2;
}

message UnsignedCertificate {
    string cid = 1;
    string unique_id = 2;
    string oracle_address = 3;
    uint64 deal_id = 4;
    string provider_address = 5;
    string data_hash = 6;
}
```

Using the `Certificate`, provider can submit consent to provide the data.

```proto
message MsgSubmitConsent {
  Consent consent = 1;
}

message Consent {
  Certificate certificate = 1;
}
```

When this consent is submitted, blockchain will check:
- if the data is provided by the owner of the data
- if the data is validated by a registered and active oracle
- if the data is provided in duplicate

If all checks pass, rewards are distributed to the provider and oracle(more about [incentive](./6-incentives.md)).

## Backwards Compatibility

Not applicable.

## Forwards Compatibility

Not applicable.

## Example Implementations

Coming soon.

## Other Implementations

None at present.

## History

- 2023-01-04: Initial draft finished

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).