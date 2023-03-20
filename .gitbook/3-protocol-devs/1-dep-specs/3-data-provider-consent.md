# Data Provider Consent

- Status: Draft
- Crated: 2023-01-04
- Modified: 2023-03-14
- Authors
  - Youngjoon Lee <yjlee@medibloc.org>
  - Gyuguen Jang <gyuguen.jang@medibloc.org>
  - Hansol Lee <hansol@medibloc.org>
  - Myongsik Gong <myongsik_gong@medibloc.org>
  - Inchul Song <icsong@medibloc.org>
  - Tae Jin Yoon <tj@medibloc.org>


## Synopsis

This document is about providing data through [DEP](../../1-users/3-data-exchange/0-about-dep.md).
Before data providers provide their data to data consumers, the data must be validated by oracle whether it matches the data type specified in the deal.

Data providers can send a request for validation to one of the oracles registered in Panacea, and the oracle will validate the data using confidential computing without any exposure to the data.
As a result of the data validation, oracle will issue a certificate to the provider.
Providers can consent to provide their data by submitting the certificate to Panacea, and will be rewarded by providing the data.
In all of these processes, the data is transmitted with encryption and stored off-chain.

### Motivation

Data should be provided by providers based on their data ownership, and the reward should be distributed in transparent and fair manner.
To do so, providers use digital signature as consent to provide the data.

### Definitions

- `Data Provider`, `Data Consumer`, and `Oracle` are defined in [User Flow](./1-user-flow.md)
- `Deal` is defined in [Data Deal](./2-data-deal.md)
- `Certificate`: a certificate that the data is validated to be provided to the deal, which is issued by oracle.

## Technical Specification

Before a data provider provides data to a deal, the data should be validated by the oracle.
If the data is successfully validated, the provider will receive a `Certificate` like below (more about [Data Validation](./4-data-validation.md)):

```proto
message Certificate {
  UnsignedCertificate unsigned_certificate = 1;
  bytes signature = 2;
}

message UnsignedCertificate {
    string unique_id = 1;
    string oracle_address = 2;
    uint64 deal_id = 3;
    string provider_address = 4;
    string data_hash = 5;
}
```

Using the `Certificate`, the data provider can submit a consent to provide the data with agreement of terms.

```proto
message MsgSubmitConsent {
  Consent consent = 1;
}

message Consent {
  uint64 deal_id = 1;
  Certificate certificate = 2;
  repeated Agreement agreements = 3;
}

message Agreement {
  uint32 term_id = 1;
  bool agreement = 2;
}
```

When this consent is submitted, blockchain will check:
- if the data is provided by the owner of the data
- if the data is validated by a registered and active oracle
- if the data is provided in duplicate
- if the provider agrees to the required terms of agreement

If all checks pass, rewards are distributed to the data provider and the oracle(more about [incentive](./6-incentives.md)).

At the end of the process, certificate is stored in `Consent`, then the data consumer can gt the data decryption key from the oracle.

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
- 2023-03-14: Delete cid from UnsignedCertificate

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).
