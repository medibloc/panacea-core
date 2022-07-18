package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRegisterOracle{}

func (msg *MsgRegisterOracle) Route() string {
	return RouterKey
}

func (msg *MsgRegisterOracle) Type() string {
	return "RegisterOracle"
}

func (msg *MsgRegisterOracle) ValidateBasic() error {
	panic("implemenets me")
}

func (msg *MsgRegisterOracle) GetSignBytes() []byte {
	panic("implemenets me")
}

func (msg *MsgRegisterOracle) GetSigners() []sdk.AccAddress {
	panic("implemenets me")
}

var _ sdk.Msg = &MsgVoteOracleRegistration{}

func (msg *MsgVoteOracleRegistration) Route() string {
	return RouterKey
}

func (msg *MsgVoteOracleRegistration) Type() string {
	return "VoteOracleRegistration"
}

func (msg *MsgVoteOracleRegistration) ValidateBasic() error {
	panic("implemenets me")
}

func (msg *MsgVoteOracleRegistration) GetSignBytes() []byte {
	panic("implemenets me")
}

func (msg *MsgVoteOracleRegistration) GetSigners() []sdk.AccAddress {
	panic("implemenets me")
}
