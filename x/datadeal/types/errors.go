package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datadeal module sentinel errors
var (
	ErrDealNotInitialized = sdkerrors.Register(ModuleName, 1, "deal has not been initialized")
	ErrDealAlreadyExist   = sdkerrors.Register(ModuleName, 2, "deal already exist")
	ErrDealNotFound       = sdkerrors.Register(ModuleName, 3, "deal is not found")
	//ErrInvalidGenesisDeal   = sdkerrors.Register(ModuleName, 4, "invalid genesis state of deal")
	ErrGetDeal          = sdkerrors.Register(ModuleName, 5, "error while get deal")
	ErrSellData         = sdkerrors.Register(ModuleName, 6, "error while selling a data")
	ErrGetDataSale      = sdkerrors.Register(ModuleName, 7, "error while get data sale")
	ErrDataSaleNotFound = sdkerrors.Register(ModuleName, 8, "data sale not found")
	//ErrDataVerificationVote = sdkerrors.Register(ModuleName, 9, "error while voting for DataVerification")
	ErrDataDeliveryVote = sdkerrors.Register(ModuleName, 10, "error while voting for DataDelivery")
)
