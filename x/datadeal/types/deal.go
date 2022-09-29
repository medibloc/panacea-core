package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const NonceSize = 12

func NewDeal(dealID uint64, msg *MsgCreateDeal) *Deal {

	dealAddress := NewDealAddress(dealID)
	nonce := make([]byte, 12)
	return &Deal{
		Id:           dealID,
		Address:      dealAddress.String(),
		DataSchema:   msg.DataSchema,
		Budget:       msg.Budget,
		MaxNumData:   msg.MaxNumData,
		CurNumData:   0,
		BuyerAddress: msg.BuyerAddress,
		Status:       DEAL_STATUS_ACTIVE,
		Nonce:        nonce,
	}
}

func NewDealAddress(dealID uint64) sdk.AccAddress {
	dealKey := "deal" + strconv.FormatUint(dealID, 10)
	return authtypes.NewModuleAddress(dealKey)
}

func (m Deal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.BuyerAddress); err != nil {
		return sdkerrors.Wrapf(err, "buyer address is invalid. address: %s", m.BuyerAddress)
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

	if m.CurNumData <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "CurNumData should be bigger than 0")
	}

	if m.CurNumData > m.MaxNumData {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "CurNumData can not be bigger than MaxNumData")
	}

	if len(m.Nonce) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "nonce is empty")
	} else if len(m.Nonce) != NonceSize {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "nonce length must be %v", NonceSize)
	}

	return nil
}
