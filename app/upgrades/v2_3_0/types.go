package v2_3_0

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

const UpgradeName = "v2.3.0"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{pnfttypes.StoreKey},
	},
}
