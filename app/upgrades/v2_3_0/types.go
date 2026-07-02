package v2_3_0

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
)

const UpgradeName = "v2.3.0"

// Upgrade wires the ICS-29 fee middleware and packet-forward-middleware stores.
// Both modules are new, so their stores are added and their default genesis is
// applied by RunMigrations during the upgrade.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			ibcfeetypes.StoreKey,
			packetforwardtypes.StoreKey,
		},
	},
}
