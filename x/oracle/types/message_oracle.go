package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterOracle{}

func NewMsgRegisterOracle(uniqueID, oracleAddress, endpoint string, nodePubKey, nodePubKeyRemoteReport []byte, trustedBlockHeight int64, trustedBlockHash []byte, oracleCommissionRate sdk.Dec) *MsgRegisterOracle {
	return &MsgRegisterOracle{
		UniqueId:               uniqueID,
		OracleAddress:          oracleAddress,
		NodePubKey:             nodePubKey,
		NodePubKeyRemoteReport: nodePubKeyRemoteReport,
		TrustedBlockHeight:     trustedBlockHeight,
		TrustedBlockHash:       trustedBlockHash,
		Endpoint:               endpoint,
		OracleCommissionRate:   oracleCommissionRate,
	}
}

func (m *MsgRegisterOracle) Route() string {
	return RouterKey
}

func (m *MsgRegisterOracle) Type() string {
	return "RegisterOracle"
}

func (m *MsgRegisterOracle) ValidateBasic() error {

	return nil
}

func (m *MsgRegisterOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgRegisterOracle) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(m.OracleAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
}

var _ sdk.Msg = &MsgApproveOracleRegistration{}

func NewMsgApproveOracleRegistration(approve *ApproveOracleRegistration, signature []byte) *MsgApproveOracleRegistration {
	return &MsgApproveOracleRegistration{
		ApproveOracleRegistration: approve,
		Signature:                 signature,
	}
}

func (m *MsgApproveOracleRegistration) Route() string {
	return RouterKey
}

func (m *MsgApproveOracleRegistration) Type() string {
	return "ApproveOracleRegistration"
}

func (m *MsgApproveOracleRegistration) ValidateBasic() error {
	if err := m.ApproveOracleRegistration.ValidateBasic(); err != nil {
		return err
	}

	if m.Signature == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature is empty")
	}

	return nil
}

func (m *MsgApproveOracleRegistration) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgApproveOracleRegistration) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(m.ApproveOracleRegistration.ApproverOracleAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
}

func (m *ApproveOracleRegistration) ValidateBasic() error {
	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueId is empty")
	}

	if len(m.ApproverOracleAddress) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "approverOracleAddress is empty")
	}

	if len(m.TargetOracleAddress) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "targetOracleAddress is empty")
	}

	if m.EncryptedOraclePrivKey == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "encryptedOraclePrivKey is empty")
	}

	return nil
}

var _ sdk.Msg = &MsgUpdateOracleInfo{}

func NewMsgUpdateOracleInfo(address, endpoint string, commissionRate sdk.Dec) *MsgUpdateOracleInfo {
	return &MsgUpdateOracleInfo{
		OracleAddress:        address,
		Endpoint:             endpoint,
		OracleCommissionRate: commissionRate,
	}
}

func (m *MsgUpdateOracleInfo) Route() string {
	return RouterKey
}

func (m *MsgUpdateOracleInfo) Type() string {
	return "UpdateOracleInfo"
}

func (m *MsgUpdateOracleInfo) ValidateBasic() error {
	if len(m.OracleAddress) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleAddress is empty")
	}

	return nil
}

func (m *MsgUpdateOracleInfo) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgUpdateOracleInfo) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(m.OracleAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
}
