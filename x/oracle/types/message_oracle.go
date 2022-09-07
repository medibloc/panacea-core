package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterOracle{}

func NewMsgRegisterOracle(uniqueID, oracleAddress string, nodePubKey, nodePubKeyRemoteReport []byte, trustedBlockHeight int64, trustedBlockHash []byte) *MsgRegisterOracle {
	return &MsgRegisterOracle{
		UniqueId:               uniqueID,
		OracleAddress:          oracleAddress,
		NodePubKey:             nodePubKey,
		NodePubKeyRemoteReport: nodePubKeyRemoteReport,
		TrustedBlockHeight:     trustedBlockHeight,
		TrustedBlockHash:       trustedBlockHash,
	}
}

func (msg *MsgRegisterOracle) Route() string {
	return RouterKey
}

func (msg *MsgRegisterOracle) Type() string {
	return "RegisterOracle"
}

func (msg *MsgRegisterOracle) ValidateBasic() error {
	if err := validateUniqueID(msg.UniqueId); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid unique ID")
	}
	if len(msg.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unique ID cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.OracleAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid oracle address (%s)", err)
	}
	if err := validateNodeKey(msg.NodePubKey); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := validateNodePubKeyRemoteReport(msg.NodePubKeyRemoteReport); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := validateTrustedBlockHeight(msg.TrustedBlockHeight); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := validateTrustedBlockHash(msg.TrustedBlockHash); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}

func (msg *MsgRegisterOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterOracle) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(msg.OracleAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
}

var _ sdk.Msg = &MsgVoteOracleRegistration{}

func (msg *MsgVoteOracleRegistration) Route() string {
	return RouterKey
}

func (msg *MsgVoteOracleRegistration) Type() string {
	return "VoteOracleRegistration"
}

func (msg *MsgVoteOracleRegistration) ValidateBasic() error {
	if msg.OracleRegistrationVote == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleRegistrationVote cannot be nil")
	}

	if len(msg.OracleRegistrationVote.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unique ID cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.OracleRegistrationVote.VoterAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid voter address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.OracleRegistrationVote.VotingTargetAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid voting target address (%s)", err)
	}
	if msg.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature cannot be nil")
	}
	return nil
}

func (msg *MsgVoteOracleRegistration) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgVoteOracleRegistration) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(msg.OracleRegistrationVote.VoterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
}
