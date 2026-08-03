package types

const (
	// ModuleName identifies the removed PNFT module for wire and upgrade
	// compatibility. It does not represent an active application module.
	ModuleName = "pnft"

	// StoreKey identifies the permanently reserved legacy PNFT KV store. The
	// v2.3 application keeps it mounted as an inactive tombstone, but no module
	// or keeper may consume or reuse its state.
	StoreKey = ModuleName
)
