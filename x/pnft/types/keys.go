package types

const (
	// ModuleName identifies the removed PNFT module for wire and upgrade
	// compatibility. It does not represent an active application module.
	ModuleName = "pnft"

	// StoreKey identifies the legacy PNFT KV store removed by the v2.3 upgrade.
	StoreKey = ModuleName
)
