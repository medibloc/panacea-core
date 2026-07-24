# NFT

Panacea v2.3 provides a policy-aware NFT module built on the standard Cosmos
SDK NFT data model. It replaces the legacy PNFT module with standard class,
NFT, owner, supply, message, query, and event contracts plus Panacea lifecycle
and policy records.

## Classes and policies

A class ID is generated as `<creator-address>:<local-class-id>`. Class metadata
and the following policies are fixed when the class is created:

- transfer policy: locked or owner-transferable;
- whether the controller may revoke an NFT;
- maximum lifetime supply, where zero means unlimited.

The creator is the initial controller. The controller may be transferred to
another non-module account, but the creator and class policies do not change.

## NFT lifecycle

The class controller mints each NFT to a recipient. An NFT ID can only be used
once within its class, including after burn.

- `send` transfers an owner-transferable NFT between non-module accounts;
- `revoke` irreversibly marks a revocable NFT as revoked without changing its
  owner or live supply;
- `burn` permanently removes the standard NFT and records a tombstone containing
  its mint, optional revocation, and burn provenance.

NFT metadata is immutable. The optional protobuf `Any` data field accepts only
`/panacea.nft.v1.BasicNFTData`.

## Query APIs

Panacea exposes both the standard `cosmos.nft.v1beta1.Query` service and the
`panacea.nft.v1.Query` service. Standard queries return live standard NFT
state. Panacea record queries additionally return policy, lifecycle, and
permanent burn tombstone information.

List queries require bounded pagination. The maximum page size is 100.

## Legacy PNFT compatibility

The `panacea.pnft.v2` module, store, REST and gRPC queries, and `pnft` CLI
commands are removed in v2.3. Legacy PNFT state is not migrated to the new NFT
stores.

Historical PNFT protobuf message types remain registered only so old
transactions and existing governance, group, or authorization records can be
decoded and exported. Executing a legacy PNFT message after the upgrade fails;
decode compatibility does not provide a PNFT runtime service.

See [Interaction with the network: CLI](../guide/interaction-with-the-network-cli.md#nft)
for current transaction and query commands.
