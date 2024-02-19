package v2_0_7

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          "v2.0.7",
	CreateUpgradeHandler: CreateUpgradeHandle,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
