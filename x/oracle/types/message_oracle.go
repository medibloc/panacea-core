package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterOracle{}

func NewMsgRegisterOracle(oracle *Oracle) *MsgRegisterOracle {
	return &MsgRegisterOracle{
		OracleDetail: oracle,
	}
}

func (msg *MsgRegisterOracle) Route() string {
	return RouterKey
}

func (msg *MsgRegisterOracle) Type() string {
	return "RegisterOracle"
}

func (msg *MsgRegisterOracle) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.OracleDetail.Address)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid oracle address (%s)", err)
	}

	if msg.OracleDetail.Endpoint == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty oracle endpoint URL")
	}
	return nil
}

func (msg *MsgRegisterOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterOracle) GetSigners() []sdk.AccAddress {
	oracle, err := sdk.AccAddressFromBech32(msg.OracleDetail.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracle}
}

var _ sdk.Msg = &MsgUpdateOracle{}

func NewMsgUpdateOracle(address, endpoint string) *MsgUpdateOracle {
	return &MsgUpdateOracle{
		Oracle:   address,
		Endpoint: endpoint,
	}
}

func (msg *MsgUpdateOracle) Route() string {
	return RouterKey
}

func (msg *MsgUpdateOracle) Type() string {
	return "UpdateOracle"
}

func (msg *MsgUpdateOracle) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Oracle)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid oracle address (%s)", err)
	}

	if msg.Endpoint == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty oracle endpoint URL")
	}
	return nil
}

func (msg *MsgUpdateOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateOracle) GetSigners() []sdk.AccAddress {
	oracle, err := sdk.AccAddressFromBech32(msg.Oracle)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracle}
}
