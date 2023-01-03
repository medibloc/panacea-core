package types

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterOracle{}

func NewMsgRegisterOracle(uniqueID, oracleAddress string, nodePubKey, nodePubKeyRemoteReport []byte, trustedBlockHeight int64, trustedBlockHash []byte, endpoint string, oracleCommissionRate, oracleCommissionMaxRate, oracleCommissionMaxChangeRate sdk.Dec) *MsgRegisterOracle {
	return &MsgRegisterOracle{
		UniqueId:                      uniqueID,
		OracleAddress:                 oracleAddress,
		NodePubKey:                    nodePubKey,
		NodePubKeyRemoteReport:        nodePubKeyRemoteReport,
		TrustedBlockHeight:            trustedBlockHeight,
		TrustedBlockHash:              trustedBlockHash,
		Endpoint:                      endpoint,
		OracleCommissionRate:          oracleCommissionRate,
		OracleCommissionMaxRate:       oracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: oracleCommissionMaxChangeRate,
	}
}

func (m *MsgRegisterOracle) Route() string {
	return RouterKey
}

func (m *MsgRegisterOracle) Type() string {
	return "RegisterOracle"
}

func (m *MsgRegisterOracle) ValidateBasic() error {
	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueId is empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleAddress is invalid. address: %s, error: %s", m.OracleAddress, err.Error())
	}

	if m.NodePubKey == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "node public key is empty")
	} else if _, err := btcec.ParsePubKey(m.NodePubKey, btcec.S256()); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid node public key")
	}

	if m.NodePubKeyRemoteReport == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "remote report of node public key is empty")
	}

	if m.TrustedBlockHeight <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "trusted block height must be greater than zero")
	}

	if m.TrustedBlockHash == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "trusted block hash should not be nil")
	}

	if m.OracleCommissionRate.LT(sdk.ZeroDec()) || m.OracleCommissionRate.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleCommissionRate must be between 0 and 1")
	}
	if m.OracleCommissionMaxRate.LT(sdk.ZeroDec()) || m.OracleCommissionMaxRate.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "OracleCommissionMaxRate must be between 0 and 1")
	}
	if m.OracleCommissionMaxChangeRate.LT(sdk.ZeroDec()) || m.OracleCommissionMaxChangeRate.GT(m.OracleCommissionMaxRate) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "OracleCommissionMaxChangeRate must be between 0 and OracleCommissionMaxRate")
	}

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

	if _, err := sdk.AccAddressFromBech32(m.ApproverOracleAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "approverOracleAddress is invalid. address: %s, error: %s", m.ApproverOracleAddress, err.Error())
	}

	if _, err := sdk.AccAddressFromBech32(m.TargetOracleAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "targetOracleAddress is invalid. address: %s, error: %s", m.TargetOracleAddress, err.Error())
	}

	if m.EncryptedOraclePrivKey == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "encryptedOraclePrivKey is empty")
	}

	return nil
}

var _ sdk.Msg = &MsgUpdateOracleInfo{}

func NewMsgUpdateOracleInfo(address, endpoint string, commissionRate *sdk.Dec) *MsgUpdateOracleInfo {
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
	if _, err := sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleAddress is invalid. address: %s, error: %s", m.OracleAddress, err.Error())
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

func (m *MsgUpgradeOracle) Route() string {
	return RouterKey
}

func (m *MsgUpgradeOracle) Type() string {
	return "UpgradeOracle"
}

func (m *MsgUpgradeOracle) ValidateBasic() error {
	// TODO: Implementation
	return nil
}

func (m *MsgUpgradeOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgUpgradeOracle) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(m.OracleAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
}

func (m *MsgApproveOracleUpgrade) Route() string {
	return RouterKey
}

func (m *MsgApproveOracleUpgrade) Type() string {
	return "ApproveOracleUpgrade"
}

func (m *MsgApproveOracleUpgrade) ValidateBasic() error {
	// TODO: Implementation
	return nil
}

func (m *MsgApproveOracleUpgrade) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgApproveOracleUpgrade) GetSigners() []sdk.AccAddress {
	oracleAddress, err := sdk.AccAddressFromBech32(m.ApproverOracleAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{oracleAddress}
}
