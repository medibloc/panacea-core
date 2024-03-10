package types

import "cosmossdk.io/errors"

var (
	ErrCreateDenom   = errors.Register(ModuleName, 1, "Error creating denom")
	ErrUpdateDenom   = errors.Register(ModuleName, 2, "Error updating denom")
	ErrDeleteDenom   = errors.Register(ModuleName, 3, "Error deleting denom")
	ErrTransferDenom = errors.Register(ModuleName, 4, "Error transfer denom")
	ErrGetDenom      = errors.Register(ModuleName, 5, "Error get denom")
	ErrMintPNFT      = errors.Register(ModuleName, 6, "Error minuting pnft")
	ErrTransferPNFT  = errors.Register(ModuleName, 7, "Error transfer pnft")
	ErrBurnPNFT      = errors.Register(ModuleName, 8, "Error burn pnft")
)
