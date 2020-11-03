# Changelog

## Unreleased

TBD

## [v1.3.2](https://github.com/medibloc/panacea-core/releases/tag/v1.3.2) - 2020-11-04

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


## [v1.2.6-internal](https://github.com/medibloc/panacea-core/releases/tag/v1.2.6-internal) - 2020-10-14

### Features

- [\#63](https://github.com/medibloc/panacea-core/pull/63) Add a new option: `halt-height` and Upgrade `medibloc/cosmos-sdk` to [v0.35.7-internal](https://github.com/medibloc/cosmos-sdk/releases/tag/v0.35.7-internal).
    - This feature will be introduced from the cosmos-sdk [v0.36.0](https://github.com/cosmos/cosmos-sdk/pull/4059).

