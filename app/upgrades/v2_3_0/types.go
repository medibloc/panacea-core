package v2_3_0

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

const UpgradeName = "v2.3.0"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added:   []string{nfttypes.StoreKey, nfttypes.PolicyStoreKey},
		Deleted: []string{pnfttypes.StoreKey},
	},
}
