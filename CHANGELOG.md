# Changelog

## Unreleased

TBD


## [v2.0.3](https://github.com/medibloc/panacea-core/releases/tag/v2.0.3) - 2022-04-06

### Improvements

- [\#291](https://github.com/medibloc/panacea-core/pull/291) Use [medibloc/cosmos-sdk@v0.42.11-panacea.1](https://github.com/medibloc/cosmos-sdk/releases/tag/v0.42.11-panacea.1) for `min_commission_rate` + add upgrade handler which sets `min_commission_rate` to 3%


### Bug fixes

- [\#213](https://github.com/medibloc/panacea-core/pull/213) Bump cosmos-sdk to [v0.42.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.9) which fixes a bug that prohibit IBC to create new channels.

### Improvements

- [\#236](https://github.com/medibloc/panacea-core/pull/236) Add an empty upgrade handler for v2.0.2.


## [v2.0.1](https://github.com/medibloc/panacea-core/releases/tag/v2.0.1) - 2021-07-28

### Bug fixes

- (x/did, x/token) [\#192](https://github.com/medibloc/panacea-core/pull/192) Return gRPC status code from `grpc_query_*.go` in `x/did` and `x/token`
- (x/token) [\#194](https://github.com/medibloc/panacea-core/pull/194) Remove legacy REST API of `x/token`
- [\#197](https://github.com/medibloc/panacea-core/pull/197) Add major version `v2` to `go.mod`
- [\#202](https://github.com/medibloc/panacea-core/pull/202) Discard the `protodep`


## [v2.0.0](https://github.com/medibloc/panacea-core/releases/tag/v2.0.0) - 2021-07-13

### Features

- (x/did) [\#97](https://github.com/medibloc/panacea-core/pull/97) Support more key types for DID verification method
- (x/did) [\#98](https://github.com/medibloc/panacea-core/pull/98) Support service / controller / various verificationRelationship in DID Document
- (x/burn) [\#115](https://github.com/medibloc/panacea-core/pull/115) Add `x/burn` module
- (stargate) [\#116](https://github.com/medibloc/panacea-core/pull/116) Support the new Cosmos SDK v0.42 Stargate
  - This PR was merged without being squashed in order to keep the full history. Detail commits can be found on the `master` branch from `36a14b3` to `b275c5e`.
  - After that, the subsequent changes below have been merged:
    - [\#157](https://github.com/medibloc/panacea-core/pull/157) feat: Add a Github Action to build/publish a Docker image
    - [\#168](https://github.com/medibloc/panacea-core/pull/168) feat: Add `java_multiple_files` option in proto files option in proto files option in proto files option in proto files
    - [\#166](https://github.com/medibloc/panacea-core/pull/166) fix: Add `make proto-lint` and Use snake-case for proto field names in order to import `genesis.json` properly
      - [\#171](https://github.com/medibloc/panacea-core/pull/171) fix: Use snake-case proto field names for `did.proto` as well
    - [\#169](https://github.com/medibloc/panacea-core/pull/169) fix: Use the protobuf `oneof` keyword for `VerificationRelationship` in the `x/did`
    - [\#174](https://github.com/medibloc/panacea-core/pull/174) fix: Accept a `did_base64` (instead of `did`) for querying a DID via REST
    - [\#177](https://github.com/medibloc/panacea-core/pull/177) feat: feat: Remove unnecessary `gogoproto.jsontag` in the `did.proto`
    - [\#183](https://github.com/medibloc/panacea-core/pull/183) feat: Bump cosmos-sdk to v0.42.7
    - [\#184](https://github.com/medibloc/panacea-core/pull/184) chore: Remove a legacy statik.go
    - [\#185](https://github.com/medibloc/panacea-core/pull/185) docs: Add our custom modules to Swagger UI
- (x/wasm) [\#141](https://github.com/medibloc/panacea-core/pull/141) Add `x/wasm` module
- [\#165](https://github.com/medibloc/panacea-core/pull/165) docs: Update Gitbook for Panacea v2
- [\#182](https://github.com/medibloc/panacea-core/pull/182) docs: Refine README for Panacea v2

### Bug fixes

- (x/token) [\#95](https://github.com/medibloc/panacea-core/pull/95) Fix to update the total supply of all coins when issuing new tokens
- (x/did) [\#100](https://github.com/medibloc/panacea-core/pull/100) Fix checking if VerificationRelationships are valid


## [v1.3.3](https://github.com/medibloc/panacea-core/releases/tag/v1.3.3) - 2020-12-03

### Bug fixes

- [\#83](https://github.com/medibloc/panacea-core/pull/83) Use cosmos/cosmos-sdk v0.37.14 (instead of medibloc/cosmos-sdk v0.37.15-internal)


## [v1.3.2](https://github.com/medibloc/panacea-core/releases/tag/v1.3.2) - 2020-11-13

### Features

- [\#80](https://github.com/medibloc/panacea-core/pull/80) Add the `x/token` module for issuing new tokens
- [\#81](https://github.com/medibloc/panacea-core/pull/81) Add private considerations for the DID specification

### Bug fixes

- [\#73](https://github.com/medibloc/panacea-core/pull/73) Make LCD use the latest height or the height specified in the URL query
- [\#74](https://github.com/medibloc/panacea-core/pull/74) Fix parsing keeper keys of `x/aol` and `x/did`
- [\#76](https://github.com/medibloc/panacea-core/pull/76) Return 404 when DID is not found or deactivated


## [v1.3.1](https://github.com/medibloc/panacea-core/releases/tag/v1.3.1) - 2020-10-19

### Features

- [\#68](https://github.com/medibloc/panacea-core/pull/68) Enhance the `x/aol` message validation
- [\#58](https://github.com/medibloc/panacea-core/pull/58) Upgrade `medibloc/cosmos-sdk` from [v0.35.6-internal](https://github.com/medibloc/cosmos-sdk/releases/tag/v0.35.6-internal) to [v0.37.15](https://github.com/medibloc/cosmos-sdk/releases/tag/v0.37.15).
- [\#66](https://github.com/medibloc/panacea-core/pull/66) Integrate with Gitbook

### Bug fixes

- [\#67](https://github.com/medibloc/panacea-core/pull/67) Use `documents` as the `x/did` genesis JSON key, instead of `Documents`.


## [v1.3.0-internal](https://github.com/medibloc/panacea-core/releases/tag/v1.3.0-internal) - 2020-09-29

### Features

- [\#7](https://github.com/medibloc/panacea-core/pull/7) ~ [\#49](https://github.com/medibloc/panacea-core/pull/49), [\#57](https://github.com/medibloc/panacea-core/pull/57) Support DID operations

### Bug fixes

- [\#59](https://github.com/medibloc/panacea-core/pull/59) Fix failed gentx because node is untrusted
    - This bug will be resolved after upgrading cosmos-sdk to [v0.37.11+](https://github.com/cosmos/cosmos-sdk/pull/6021).


## [v1.2.7-internal](https://github.com/medibloc/panacea-core/releases/tag/v1.2.7-internal) - 2020-12-03

### Bug fixes

- [\#82](https://github.com/medibloc/panacea-core/pull/82) Fix parsing keeper keys of `x/aol`


## [v1.2.6-internal](https://github.com/medibloc/panacea-core/releases/tag/v1.2.6-internal) - 2020-10-14

### Features

- [\#63](https://github.com/medibloc/panacea-core/pull/63) Add a new option: `halt-height` and Upgrade `medibloc/cosmos-sdk` to [v0.35.7-internal](https://github.com/medibloc/cosmos-sdk/releases/tag/v0.35.7-internal).
    - This feature will be introduced from the cosmos-sdk [v0.36.0](https://github.com/cosmos/cosmos-sdk/pull/4059).

