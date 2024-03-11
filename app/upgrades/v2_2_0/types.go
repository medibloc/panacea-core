package v2_2_0

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          "v2.2.0",
	CreateUpgradeHandler: CreateUpgradeHandle,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			consensustypes.ModuleName,
			crisistypes.ModuleName,
			group.ModuleName,
			pnfttypes.ModuleName,
		},
	},
}
