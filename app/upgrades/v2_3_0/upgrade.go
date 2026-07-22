package v2_3_0

import (
	"context"

	"cosmossdk.io/store/prefix"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/medibloc/panacea-core/v2/app/keepers"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, keepers *keepers.AppKeepersWithKey) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		toVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return fromVM, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		versionStore := prefix.NewStore(
			sdkCtx.KVStore(keepers.GetKey(upgradetypes.StoreKey)),
			[]byte{upgradetypes.VersionMapByte},
		)
		versionStore.Delete([]byte(pnfttypes.ModuleName))

		return toVM, nil
	}
}
