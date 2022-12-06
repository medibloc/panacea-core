package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
