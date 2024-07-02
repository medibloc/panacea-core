package v2_2_1

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          "v2.2.1",
	CreateUpgradeHandler: CreateUpgradeHandle,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
