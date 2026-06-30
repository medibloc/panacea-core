package v2_0_5

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/medibloc/panacea-core/v2/app/upgrades"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          "v2.0.5",
	CreateUpgradeHandler: CreateUpgradeHandle,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added:   []string{authz.ModuleName, feegrant.ModuleName},
		Deleted: []string{"token"},
	},
}
