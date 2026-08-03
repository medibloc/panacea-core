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
		Added: []string{nfttypes.StoreKey, nfttypes.PolicyStoreKey},
		// pnft is permanently reserved. In-place upgraded nodes may retain its
		// orphaned IAVL data, so reusing this key could diverge from fresh or
		// state-synced nodes.
		Deleted: []string{pnfttypes.StoreKey},
	},
}
