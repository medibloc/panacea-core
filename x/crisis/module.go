package crisis

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	parent "github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-core/types/assets"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the crisis module.
type AppModuleBasic struct{}

// Name returns the crisis module's name
func (AppModuleBasic) Name() string {
	return parent.AppModuleBasic{}.Name()
}

// RegisterCodec registers the crisis module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	parent.RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the crisis module.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	// customize to set default genesis state constant fee denum to umed
	defaultGenesisState := parent.DefaultGenesisState()
	defaultGenesisState.ConstantFee.Denom = assets.MicroMedDenom

	return parent.ModuleCdc.MustMarshalJSON(defaultGenesisState)
}

// ValidateGenesis performs genesis state validation for the crisis module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return parent.AppModuleBasic{}.ValidateGenesis(bz)
}

// RegisterRESTRoutes registers the REST routes for the crisis module.
func (AppModuleBasic) RegisterRESTRoutes(cliCtx context.CLIContext, router *mux.Router) {
	parent.AppModuleBasic{}.RegisterRESTRoutes(cliCtx, router)
}

// GetTxCmd returns the root tx command for the crisis module.
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return parent.AppModuleBasic{}.GetTxCmd(cdc)
}

// GetQueryCmd returns the root query command for the crisis module.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return parent.AppModuleBasic{}.GetQueryCmd(cdc)
}

//__________________________________________________________________________

// AppModule implements an application module for the crisis module.
type AppModule struct {
	AppModuleBasic
	sdkAppModule parent.AppModule
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper *Keeper) AppModule {
	return AppModule{
		AppModuleBasic:  AppModuleBasic{},
		sdkAppModule: NewCosmosAppModule(keeper),
	}
}

// InitGenesis performs genesis initialization for the crisis module.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return am.sdkAppModule.InitGenesis(ctx, data)
}

// ExportGenesis returns the exported genesis state as raw bytes for the crisis module.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return am.sdkAppModule.ExportGenesis(ctx)
}

// RegisterInvariants registers the crisis module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	am.sdkAppModule.RegisterInvariants(ir)
}

// Route returns the message routing key for the crisis module.
func (am AppModule) Route() string {
	return am.sdkAppModule.Route()
}

// NewHandler returns an sdk.Handler for the crisis module.
func (am AppModule) NewHandler() sdk.Handler {
	return am.sdkAppModule.NewHandler()
}

// QuerierRoute returns the crisis module's querier route name.
func (am AppModule) QuerierRoute() string {
	return am.sdkAppModule.QuerierRoute()
}

// NewQuerierHandler returns the crisis module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return am.sdkAppModule.NewQuerierHandler()
}

// BeginBlock returns the begin blocker for the crisis module.
func (am AppModule) BeginBlock(ctx sdk.Context, rbb abci.RequestBeginBlock) {
	am.sdkAppModule.BeginBlock(ctx, rbb)
}

// EndBlock returns the end blocker for the crisis module.
func (am AppModule) EndBlock(ctx sdk.Context, rbb abci.RequestEndBlock) []abci.ValidatorUpdate {
	return am.sdkAppModule.EndBlock(ctx, rbb)
}




