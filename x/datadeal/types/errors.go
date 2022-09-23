package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datadeal module sentinel errors
var (
	ErrDealNotInitialized = sdkerrors.Register(ModuleName, 1, "deal has not been initialized")
	ErrDealAlreadyExist   = sdkerrors.Register(ModuleName, 2, "deal already exist")
	ErrDealNotFound       = sdkerrors.Register(ModuleName, 3, "deal is not found")
	ErrInvalidGenesisDeal = sdkerrors.Register(ModuleName, 4, "invalid genesis state of deal")
	ErrGetDeal            = sdkerrors.Register(ModuleName, 5, "error while get deal")
)
