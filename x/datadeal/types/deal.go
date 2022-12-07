package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func NewDeal(dealID uint64, msg *MsgCreateDeal) *Deal {
	dealAddress := NewDealAddress(dealID)

	return &Deal{
		Id:              dealID,
		Address:         dealAddress.String(),
		DataSchema:      msg.DataSchema,
		Budget:          msg.Budget,
		MaxNumData:      msg.MaxNumData,
		CurNumData:      0,
		ConsumerAddress: msg.BuyerAddress,
		Status:          DEAL_STATUS_ACTIVE,
	}
}

func NewDealAddress(dealID uint64) sdk.AccAddress {
	dealKey := "deal" + strconv.FormatUint(dealID, 10)
	return authtypes.NewModuleAddress(dealKey)
}

func (m Deal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.ConsumerAddress); err != nil {
		return sdkerrors.Wrapf(err, "consumer address is invalid. address: %s", m.ConsumerAddress)
	}
	if len(m.DataSchema) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "there is no data schema")
	}
	if m.Id <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "ID should be bigger than 0")
	}
	if m.MaxNumData <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "MaxNumData should be bigger than 0")
	}

	if m.Budget == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is empty")
	}

	if !m.Budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}

	if m.CurNumData > m.MaxNumData {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "CurNumData can not be bigger than MaxNumData")
	}

	return nil
}
