package oracle

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/keeper"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	handlerOracleUpgrade(ctx, keeper)
}

func handlerOracleUpgrade(ctx sdk.Context, keeper keeper.Keeper) {
	upgradeInfo, err := keeper.GetOracleUpgradeInfo(ctx)
	if err != nil {
		if errors.Is(err, types.ErrOracleUpgradeInfoNotFound) {
			return
		} else {
			panic(err)
		}
	}

	if upgradeInfo.ValidateOracleUpgradeInfo(ctx) {
		if err := keeper.ApplyUpgrade(ctx, upgradeInfo); err != nil {
			panic(err)
		}
		keeper.RemoveOracleUpgradeInfo(ctx)
	}
}
