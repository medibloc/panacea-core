package v2_0_5

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	"github.com/medibloc/panacea-core/v2/app/keepers"
)

func CreateUpgradeHandle(mm *module.Manager, configurator module.Configurator, keepers *keepers.AppKeepersWithKey) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// 1st-time running in-store migrations, using 1 as fromVersion to
		// avoid running InitGenesis.
		fromVM := map[string]uint64{
			"auth":         1,
			"bank":         1,
			"capability":   1,
			"crisis":       1,
			"distribution": 1,
			"evidence":     1,
			"gov":          1,
			"mint":         1,
			"params":       1,
			"slashing":     1,
			"staking":      1,
			"upgrade":      1,
			"vesting":      1,
			"ibc":          1,
			"genutil":      1,
			"transfer":     1,
			"aol":          1,
			"did":          1,
			"burn":         1,
			"wasm":         1,
		}

		keepers.IBCKeeper.ConnectionKeeper.SetParams(sdkCtx, ibcconnectiontypes.DefaultParams())

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
