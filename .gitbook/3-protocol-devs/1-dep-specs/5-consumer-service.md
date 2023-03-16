# Consumer Service

- Status: Draft
- Created: 2023-03-15
- Modified: 
- Authors
    - Myongsik Gong <myongsik_gong@medibloc.org>
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Youngjoon Lee <yjlee@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Inchul Song <icsong@medibloc.org>
    - Tae Jin Yoon <tj@medibloc.org>


## Synopsis

This document describes consumer service.

### Motivation
Data consumers should run an HTTP server to receive the encrypted data from oracle.
When creating deals, data consumers must specify the endpoint of their server in the `consumer-service-endpoint` field.

When the incentive is paid to the data provider completely, the data consumer can get the `secret key` to decrypt data received from the oracle.

### Definitions
- `Data Provider`, `Data Consumer`, and `Oracle` are defined in [User Flow](./1-user-flow.md)
- `Deal` is defined in [Data Deal](./2-data-deal.md)
- `Consumer Service`: HTTP server to store data provided by oracle

## Technical Specification

### Consumer service specification
The consumer service must be able to handle the HTTP POST request below:

#### Request URI
```http request
POST /v0/deals/${dealId}/data/${dataHash}
```

#### Request Headers
```
# Authorization: Bearer {jwtToken}
# Content-Type: application/json
```

The JWT is signed by oracle using the oracle private key.
The `consumer service` can use this JWT to authenticate that oracle sent the post request.

## Backwards Compatibility

Not applicable.

## Forwards Compatibility

Not applicable.

## Example Implementations

With the following repository, you can run a simple `consumer service` that satisfies the specification.
- https://github.com/medibloc/panacea-dep-consumer

## Other Implementations

None at present.

## History

- 2023-03-15: Initial draft finished

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).