package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateDeal{}

func NewMsgCreateDeal(dataSchema []string, budget *sdk.Coin, maxNumData uint64, trustedDataValidator []string, owner string) *MsgCreateDeal {
	return &MsgCreateDeal{
		DataSchema:            dataSchema,
		Budget:                budget,
		MaxNumData:            maxNumData,
		TrustedDataValidators: trustedDataValidator,
		Owner:                 owner,
	}
}

func (msg *MsgCreateDeal) Route() string {
	return RouterKey
}

func (msg *MsgCreateDeal) Type() string {
	return "CreateDeal"
}

// ValidateBasic is validation for MsgCreateDeal.
func (msg *MsgCreateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	budget := msg.Budget
	if !budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}

	for _, validator := range msg.TrustedDataValidators {
		_, err = sdk.AccAddressFromBech32(validator)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
		}
	}
	return nil
}

func (msg *MsgCreateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
