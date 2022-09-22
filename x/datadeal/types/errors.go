package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datadeal module sentinel errors
var (
	ErrSellData         = sdkerrors.Register(ModuleName, 1, "error while selling a data")
	ErrGetDataSale      = sdkerrors.Register(ModuleName, 2, "error while get data sale")
	ErrDataSaleNotFound = sdkerrors.Register(ModuleName, 3, "data sale not found")
)
