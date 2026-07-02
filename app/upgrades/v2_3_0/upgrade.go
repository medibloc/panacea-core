package v2_3_0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/medibloc/panacea-core/v2/app/keepers"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, _ *keepers.AppKeepersWithKey) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// RunMigrations initializes the newly added 29-fee and packet-forward
		// modules with their default genesis (fee percentage 0, no in-flight packets).
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
