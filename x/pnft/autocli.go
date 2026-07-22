package pnft

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

// AutoCLIOptions disables inferred commands for the legacy PNFT services.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{}
}
