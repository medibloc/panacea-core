package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datadeal module sentinel errors

var (
	ErrDealNotInitialized  = sdkerrors.Register(ModuleName, 1, "deal has not been initialized")
	ErrDealAlreadyExist    = sdkerrors.Register(ModuleName, 2, "deal already exist")
	ErrDealNotFound        = sdkerrors.Register(ModuleName, 3, "deal is not found")
	ErrCertificateNotFound = sdkerrors.Register(ModuleName, 4, "certificate is not found")
	ErrGetCertificate      = sdkerrors.Register(ModuleName, 5, "error while get certificate")
	ErrSubmitConsent       = sdkerrors.Register(ModuleName, 6, "error while submit consent")
)
