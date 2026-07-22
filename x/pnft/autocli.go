package pnft

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

// AutoCLIOptions disables inferred commands for the legacy PNFT services.
// The remaining query commands are provided explicitly by AppModuleBasic.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{}
}
