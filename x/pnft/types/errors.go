package types

import "cosmossdk.io/errors"

var (
	ErrCreateDenom = errors.Register(ModuleName, 1, "Error creating denom")
	ErrUpdateDenom = errors.Register(ModuleName, 2, "Error updating denom")
	ErrDeleteDenom = errors.Register(ModuleName, 3, "Error deleting denom")

	ErrTransferDenom = errors.Register(ModuleName, 4, "Error transfer denom")
	ErrGetDenom      = errors.Register(ModuleName, 5, "Error get denom")
)
