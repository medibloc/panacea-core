package pan_19_noop_rehearsal

import (
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/medibloc/panacea-core/v2/app/keepers"
)

// CreateUpgradeHandler registers a temporary local-only no-op handler for PAN-19
// Cosmovisor rehearsal. PAN-30 will replace this with the real v2.3.0 upgrade.
func CreateUpgradeHandler(_ *module.Manager, _ module.Configurator, _ *keepers.AppKeepersWithKey) upgradetypes.UpgradeHandler {
	return func(_ sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return fromVM, nil
	}
}
