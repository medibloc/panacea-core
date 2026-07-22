package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	upstreamnft "cosmossdk.io/x/nft"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

// Keeper owns the canonical SDK NFT store and Panacea policy store as one
// module boundary. The embedded SDK keeper remains private so callers cannot
// bypass Panacea policy checks through the upstream write API.
type Keeper struct {
	cdc                codec.BinaryCodec
	addressCodec       address.Codec
	nftStoreService    corestore.KVStoreService
	policyStoreService corestore.KVStoreService
	accountKeeper      upstreamnft.AccountKeeper
	nftKeeper          upstreamkeeper.Keeper
	moduleAccounts     map[string]struct{}

	schema        collections.Schema
	classPolicies collections.Map[string, types.ClassPolicy]
	mintedCounts  collections.Map[string, uint64]
	lifecycles    collections.Map[collections.Pair[string, string], types.LifecycleRecord]
	tombstones    collections.Map[collections.Pair[string, string], types.BurnTombstone]
}

// NewKeeper creates the single keeper that owns both NFT stores.
func NewKeeper(
	cdc codec.BinaryCodec,
	nftStoreService corestore.KVStoreService,
	policyStoreService corestore.KVStoreService,
	accountKeeper upstreamnft.AccountKeeper,
	bankKeeper upstreamnft.BankKeeper,
	moduleAccountAddresses []sdk.AccAddress,
) Keeper {
	if cdc == nil {
		panic("nft keeper requires a codec")
	}
	if nftStoreService == nil {
		panic("nft keeper requires the nft store service")
	}
	if policyStoreService == nil {
		panic("nft keeper requires the panacea_nft store service")
	}
	if accountKeeper == nil {
		panic("nft keeper requires an account keeper")
	}
	if bankKeeper == nil {
		panic("nft keeper requires a bank keeper")
	}
	if len(moduleAccountAddresses) == 0 {
		panic("nft keeper requires module account addresses")
	}
	addressCodec := accountKeeper.AddressCodec()
	if addressCodec == nil {
		panic("nft keeper requires an account address codec")
	}
	moduleAccounts := make(map[string]struct{}, len(moduleAccountAddresses))
	for _, moduleAddress := range moduleAccountAddresses {
		if len(moduleAddress) == 0 {
			panic("nft keeper requires non-empty module account addresses")
		}
		moduleAccounts[string(moduleAddress)] = struct{}{}
	}

	schemaBuilder := collections.NewSchemaBuilder(policyStoreService)
	classPolicies := collections.NewMap(
		schemaBuilder,
		classPoliciesPrefix,
		"class_policies",
		collections.StringKey,
		codec.CollValue[types.ClassPolicy](cdc),
	)
	mintedCounts := collections.NewMap(
		schemaBuilder,
		mintedCountsPrefix,
		"minted_counts",
		collections.StringKey,
		collections.Uint64Value,
	)
	lifecycles := collections.NewMap(
		schemaBuilder,
		lifecyclesPrefix,
		"lifecycles",
		nftKeyCodec,
		codec.CollValue[types.LifecycleRecord](cdc),
	)
	tombstones := collections.NewMap(
		schemaBuilder,
		tombstonesPrefix,
		"tombstones",
		nftKeyCodec,
		codec.CollValue[types.BurnTombstone](cdc),
	)
	schema, err := schemaBuilder.Build()
	if err != nil {
		panic(fmt.Errorf("build panacea_nft schema: %w", err))
	}

	return Keeper{
		cdc:                cdc,
		addressCodec:       addressCodec,
		nftStoreService:    nftStoreService,
		policyStoreService: policyStoreService,
		accountKeeper:      accountKeeper,
		moduleAccounts:     moduleAccounts,
		nftKeeper: upstreamkeeper.NewKeeper(
			nftStoreService,
			cdc,
			accountKeeper,
			bankKeeper,
		),
		schema:        schema,
		classPolicies: classPolicies,
		mintedCounts:  mintedCounts,
		lifecycles:    lifecycles,
		tombstones:    tombstones,
	}
}

// Logger returns a module-scoped logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
