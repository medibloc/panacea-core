package nft

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	upstreamnft "cosmossdk.io/x/nft"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/medibloc/panacea-core/v2/x/nft/keeper"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.HasABCIGenesis      = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasServices         = AppModule{}
	_ appmodule.AppModule        = AppModule{}
)

// AppModuleBasic implements the stateless NFT module wiring.
type AppModuleBasic struct {
	addressCodec address.Codec
}

// NewAppModuleBasic creates the basic NFT module.
func NewAppModuleBasic(addressCodec address.Codec) AppModuleBasic {
	if addressCodec == nil {
		panic("nft module requires an address codec")
	}
	return AppModuleBasic{addressCodec: addressCodec}
}

// Name returns the runtime module name shared with the standard NFT store.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec intentionally registers no NFT types. NFT legacy
// JSON signing is handled by the protobuf-based v0.50 sign mode handler.
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces delegates all standard and Panacea wire registration to
// the types package without registering the upstream NFT AppModule.
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns a non-nil empty standard NFT genesis.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validates the combined standard and policy genesis contract.
func (am AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, data json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(data, &genesis); err != nil {
		return fmt.Errorf("unmarshal %s genesis: %w", types.ModuleName, err)
	}
	unpacker, ok := cdc.(cdctypes.AnyUnpacker)
	if !ok {
		return fmt.Errorf("%s genesis codec cannot unpack protobuf Any values", types.ModuleName)
	}
	return types.ValidateGenesis(genesis, am.addressCodec, unpacker)
}

// RegisterGRPCGatewayRoutes registers the standard and Panacea query routes.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientContext client.Context, mux *runtime.ServeMux) {
	if err := upstreamnft.RegisterQueryHandlerClient(
		context.Background(),
		mux,
		upstreamnft.NewQueryClient(clientContext),
	); err != nil {
		panic(err)
	}
	if err := types.RegisterQueryHandlerClient(
		context.Background(),
		mux,
		types.NewQueryClient(clientContext),
	); err != nil {
		panic(err)
	}
}

// AppModule is Panacea's single NFT runtime module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates the NFT AppModule around the integrated keeper.
func NewAppModule(addressCodec address.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(addressCodec),
		keeper:         k,
	}
}

// IsOnePerModuleType implements depinject.OnePerModuleType.
func (AppModule) IsOnePerModuleType() {}

// IsAppModule marks AppModule as a core app module.
func (AppModule) IsAppModule() {}

// QuerierRoute returns the legacy query route name.
func (AppModule) QuerierRoute() string { return types.QuerierRoute }

// RegisterServices binds both message and query APIs to Panacea's policy-aware
// servers. The unrestricted upstream NFT keeper servers remain unreachable.
func (am AppModule) RegisterServices(configurator module.Configurator) {
	types.RegisterMsgServer(configurator.MsgServer(), keeper.NewMsgServer(am.keeper))
	upstreamnft.RegisterMsgServer(configurator.MsgServer(), keeper.NewStandardMsgServer(am.keeper))
	types.RegisterQueryServer(configurator.QueryServer(), keeper.NewQueryServer(am.keeper))
	upstreamnft.RegisterQueryServer(
		configurator.QueryServer(),
		keeper.NewStandardQueryServer(am.keeper),
	)
}

// RegisterInvariants registers no runtime-wide invariant route.
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

// InitGenesis atomically initializes both stores and returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesis)
	if err := am.keeper.InitGenesis(ctx, &genesis); err != nil {
		panic(err)
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis deterministically exports the combined state.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(genesis)
}

// ConsensusVersion returns the initial combined module version.
func (AppModule) ConsensusVersion() uint64 { return 1 }
