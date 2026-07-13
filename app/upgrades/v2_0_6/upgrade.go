package v2_0_6

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/medibloc/panacea-core/v2/app/keepers"
)

func CreateUpgradeHandle(mm *module.Manager, configurator module.Configurator, _ *keepers.AppKeepersWithKey) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// transfer module consensus version has been bumped to 2
		// https://ibc.cosmos.network/main/migrations/v3-to-v4.html#migration-to-fix-support-for-base-denoms-with-slashes
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
