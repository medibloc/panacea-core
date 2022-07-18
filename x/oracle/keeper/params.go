package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

var (
	params = types.Params{}
)

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	params = p
}
