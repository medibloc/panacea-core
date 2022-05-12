package keeper

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/CosmWasm/wasmd/x/wasm"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"

	// this line is used by starport scaffolding # ibc/keeper/import

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc      codec.Marshaler
		storeKey sdk.StoreKey
		memKey   sdk.StoreKey

		paramSpace paramtypes.Subspace

		// keepers
		bankKeeper    types.BankKeeper
		accountKeeper types.AccountKeeper
		wasmKeeper    wasmtypes.ContractOpsKeeper
		viewKeeper    wasmtypes.ViewKeeper
	}
)

var _ wasmtypes.ViewKeeper = (*wasm.Keeper)(nil)

func NewKeeper(
	cdc codec.Marshaler,
	storeKey,
	memKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	wasmKeeper wasm.Keeper,
) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramSpace:    paramSpace,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		wasmKeeper:    wasmkeeper.NewDefaultPermissionKeeper(wasmKeeper),
		viewKeeper:    wasmKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
