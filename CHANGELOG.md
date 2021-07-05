# Changelog

## Unreleased

TBD


## [v2.0.0-alpha.1](https://github.com/medibloc/panacea-core/releases/tag/v2.0.0-alpha.1) - 2021-07-01

### Features

- [\#116](https://github.com/medibloc/panacea-core/pull/116) feat: Support the new Cosmos SDK v0.42 Stargate
- [\#141](https://github.com/medibloc/panacea-core/pull/141) feat: Add `x/wasm` module
- [\#115](https://github.com/medibloc/panacea-core/pull/115) feat: Add `x/burn` module

### Bug fixes

- [\#95](https://github.com/medibloc/panacea-core/pull/95) x/token: Fix to update the total supply of all coins when issuing new tokens


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

