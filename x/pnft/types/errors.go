package types

import "cosmossdk.io/errors"

var (
	ErrCreateDenom   = errors.Register(ModuleName, 1, "failed to create denom")
	ErrUpdateDenom   = errors.Register(ModuleName, 2, "failed to update denom")
	ErrDeleteDenom   = errors.Register(ModuleName, 3, "failed to delete denom")
	ErrTransferDenom = errors.Register(ModuleName, 4, "failed to transfer denom")
	ErrGetDenom      = errors.Register(ModuleName, 5, "failed to get denom")
	ErrMintPNFT      = errors.Register(ModuleName, 6, "failed to mint pnft")
	ErrTransferPNFT  = errors.Register(ModuleName, 7, "failed to transfer pnft")
	ErrBurnPNFT      = errors.Register(ModuleName, 8, "failed to burn pnft")
	ErrGetPNFT       = errors.Register(ModuleName, 9, "failed to get pnft")
)
