package v2_3_0

import (
	"context"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/medibloc/panacea-core/v2/app/keepers"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, keepers *keepers.AppKeepersWithKey) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		toVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return fromVM, err
		}

		if fromVersion, exists := fromVM[govtypes.ModuleName]; !exists || fromVersion >= toVM[govtypes.ModuleName] {
			return toVM, nil
		}

		params, err := keepers.GovKeeper.Params.Get(ctx)
		if err != nil {
			return fromVM, err
		}
		params.ExpeditedMinDeposit = sdk.Coins(params.MinDeposit).MulInt(
			math.NewInt(govv1.DefaultMinExpeditedDepositTokensRatio),
		)
		if err := keepers.GovKeeper.Params.Set(ctx, params); err != nil {
			return fromVM, err
		}

		return toVM, nil
	}
}
