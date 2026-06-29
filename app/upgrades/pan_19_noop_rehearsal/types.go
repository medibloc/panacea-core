package pan_19_noop_rehearsal

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
)

const UpgradeName = "pan-19-noop-rehearsal"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
