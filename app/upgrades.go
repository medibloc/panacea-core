package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (app App) RegisterUpgradeHandlers() {
	// Set param key table for params module migration
	for _, subspace := range app.ParamsKeeper.GetSubspaces() {
		subspace := subspace

		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable() //nolint:staticcheck
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}

		baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

		app.UpgradeKeeper.SetUpgradeHandler(
			"v2.2.0",
			func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
				// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
				baseapp.MigrateParams(ctx, baseAppLegacySS, &app.ConsensusParamsKeeper)

				// Note: this migration is optional,
				// You can include x/gov proposal migration documented in [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md)

				return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
			},
		)

		upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
		if err != nil {
			panic(err)
		}

		if upgradeInfo.Name == "v2.2.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			storeUpgrades := storetypes.StoreUpgrades{
				Added: []string{
					consensustypes.ModuleName,
					crisistypes.ModuleName,
				},
			}

			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}

}

func (app App) UpradeHandler_v2_2_0() {
	upgradeName := "v2.2.0"
	app.UpgradeKeeper.SetUpgradeHandler(
		upgradeName,
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// Set param key table for params module migration
			for _, subspace := range app.ParamsKeeper.GetSubspaces() {
				subspace := subspace

				var keyTable paramstypes.KeyTable
				switch subspace.Name() {
				case authtypes.ModuleName:
					keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
				case banktypes.ModuleName:
					keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
				case stakingtypes.ModuleName:
					keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
				case minttypes.ModuleName:
					keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
				case distrtypes.ModuleName:
					keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
				case slashingtypes.ModuleName:
					keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
				case govtypes.ModuleName:
					keyTable = govv1.ParamKeyTable() //nolint:staticcheck
				case crisistypes.ModuleName:
					keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
				}

				if !subspace.HasKeyTable() {
					subspace.WithKeyTable(keyTable)
				}
			}

			baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

			// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
			baseapp.MigrateParams(ctx, baseAppLegacySS, &app.ConsensusParamsKeeper)

			// Note: this migration is optional,
			// You can include x/gov proposal migration documented in [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md)

			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == upgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				consensustypes.ModuleName,
				crisistypes.ModuleName,
				nft.ModuleName,
				group.ModuleName,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
