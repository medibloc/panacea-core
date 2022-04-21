package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datapool module sentinel errors
var (
	ErrDataValidatorNotFound      = sdkerrors.Register(ModuleName, 1, "data validator not found")
	ErrDataValidatorAlreadyExist  = sdkerrors.Register(ModuleName, 2, "data validator already exists")
	ErrNotEnoughPoolDeposit       = sdkerrors.Register(ModuleName, 3, "the deposit is not enough to make a data pool")
	ErrNotRegisteredDataValidator = sdkerrors.Register(ModuleName, 4, "data validator is not registered")
	ErrNoRegisteredNFTContract    = sdkerrors.Register(ModuleName, 5, "no NFT contract is registered")
	ErrPoolNotFound               = sdkerrors.Register(ModuleName, 6, "pool not found")
	ErrNotEqualsSeller            = sdkerrors.Register(ModuleName, 7, "the requester does not matched certificate information")
	ErrInvalidSignature           = sdkerrors.Register(ModuleName, 8, "failed to signature verify.")
	ErrInvalidDataValidationCert  = sdkerrors.Register(ModuleName, 9, "certificate is not valid")
	ErrExistSameDataHash          = sdkerrors.Register(ModuleName, 10, "data already exists in the pool")
	ErrGetDataValidationCert      = sdkerrors.Register(ModuleName, 11, "failed get certificate.")
	ErrFailedMintShareToken       = sdkerrors.Register(ModuleName, 12, "failed mint share token.")
)
