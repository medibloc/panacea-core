package keeper

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"

	// this line is used by starport scaffolding # ibc/keeper/import

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type (
	Keeper struct {
		cdc           codec.Marshaler
		storeKey      sdk.StoreKey
		memKey        sdk.StoreKey
		bankKeeper    types.BankKeeper
		accountKeeper types.AccountKeeper
		wasmKeeper    wasmtypes.ContractOpsKeeper
	}
)

func NewKeeper(
	cdc codec.Marshaler,
	storeKey,
	memKey sdk.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	wasmKeeper wasm.Keeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		wasmKeeper:    wasmkeeper.NewDefaultPermissionKeeper(wasmKeeper),
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
