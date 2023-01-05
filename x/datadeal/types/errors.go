package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datadeal module sentinel errors

var (
	ErrDealNotInitialized = sdkerrors.Register(ModuleName, 1, "deal has not been initialized")
	ErrDealAlreadyExist   = sdkerrors.Register(ModuleName, 2, "deal already exist")
	ErrDealNotFound       = sdkerrors.Register(ModuleName, 3, "deal is not found")
	ErrGetDeal            = sdkerrors.Register(ModuleName, 4, "error while get deal")
	ErrConsentNotFound    = sdkerrors.Register(ModuleName, 5, "consent is not found")
	ErrGetConsent         = sdkerrors.Register(ModuleName, 6, "error while get consent")
	ErrSubmitConsent      = sdkerrors.Register(ModuleName, 7, "error while submit consent")
	ErrDeactivateDeal     = sdkerrors.Register(ModuleName, 8, "error while deactivate deal")
)
