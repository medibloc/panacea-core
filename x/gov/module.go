package gov


import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	parent "github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-core/types/assets"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type SdkAppModuleBasic    = parent.AppModuleBasic

type AppModuleBasic struct{
	SdkAppModuleBasic
}

// NewAppModuleBasic creates a new AppModuleBasic object
func NewAppModuleBasic(proposalHandlers ...client.ProposalHandler) AppModuleBasic {
	return AppModuleBasic{
		SdkAppModuleBasic: parent.NewAppModuleBasic(proposalHandlers...),
	}
}

func (AppModuleBasic) Name() string {
	return parent.AppModuleBasic{}.Name()
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	parent.RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	defaultGenesisState := parent.DefaultGenesisState()
	defaultGenesisState.DepositParams.MinDeposit[0].Denom = assets.MicroMedDenom

	return parent.ModuleCdc.MustMarshalJSON(defaultGenesisState)
}

func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return parent.AppModuleBasic{}.ValidateGenesis(bz)
}

func (AppModuleBasic) RegisterRESTRoutes(cliCtx context.CLIContext, router *mux.Router) {
	parent.AppModuleBasic{}.RegisterRESTRoutes(cliCtx, router)
}

func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return parent.AppModuleBasic{}.GetTxCmd(cdc)
}

func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return parent.AppModuleBasic{}.GetQueryCmd(cdc)
}

//__________________________________________________________________________

type AppModule struct {
	AppModuleBasic
	sdkAppModule parent.AppModule
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper, supplyKeeper parent.SupplyKeeper) AppModule {
	return AppModule{
		AppModuleBasic:  AppModuleBasic{},
		sdkAppModule: NewCosmosAppModule(keeper, supplyKeeper),
	}
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return am.sdkAppModule.InitGenesis(ctx, data)
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return am.sdkAppModule.ExportGenesis(ctx)
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	am.sdkAppModule.RegisterInvariants(ir)
}

func (am AppModule) Route() string {
	return am.sdkAppModule.Route()
}

func (am AppModule) NewHandler() sdk.Handler {
	return am.sdkAppModule.NewHandler()
}

func (am AppModule) QuerierRoute() string {
	return am.sdkAppModule.QuerierRoute()
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return am.sdkAppModule.NewQuerierHandler()
}

func (am AppModule) BeginBlock(ctx sdk.Context, rbb abci.RequestBeginBlock) {
	am.sdkAppModule.BeginBlock(ctx, rbb)
}

func (am AppModule) EndBlock(ctx sdk.Context, rbb abci.RequestEndBlock) []abci.ValidatorUpdate {
	return am.sdkAppModule.EndBlock(ctx, rbb)
}




