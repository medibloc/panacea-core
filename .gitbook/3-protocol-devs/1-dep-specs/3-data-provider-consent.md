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

deal에 명시된 

- 올바른 데이터의 제공 (data provision abusing)
- privacy preserving
- provider의 제공 동의
- data provision에 따른 투명한 보상 분배

### Definitions

- `Data Provider`, `Data Consumer`, and `Oracle` are defined in [User Flow](./1-user-flow.md)
- `Deal`
- `confidential computing`
- `certificate`

## Technical Specification



## Backwards Compatibility

Not applicable.

## Forwards Compatibility

vp

## Example Implementations

Coming soon.

## Other Implementations

None at present.

## History

- 2023-01-04: Initial draft finished

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).