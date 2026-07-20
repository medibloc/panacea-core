package nft

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
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

// ValidateGenesis validates the empty combined genesis contract.
func (am AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, data json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(data, &genesis); err != nil {
		return fmt.Errorf("unmarshal %s genesis: %w", types.ModuleName, err)
	}
	return types.ValidateGenesis(genesis, am.addressCodec)
}

// RegisterGRPCGatewayRoutes is intentionally empty until the standard and
// Panacea query wrappers are implemented.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

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

// RegisterServices intentionally registers neither the upstream NFT services
// nor incomplete Panacea handlers in the empty skeleton.
func (AppModule) RegisterServices(module.Configurator) {}

// RegisterInvariants registers no runtime-wide invariant route.
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

// InitGenesis initializes both empty stores and returns no validator updates.
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
