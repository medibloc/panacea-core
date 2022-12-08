package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewOracle(oracleAddress, uniqueID, endpoint string, oracleCommissionRate sdk.Dec) *Oracle {
	return &Oracle{
		OracleAddress:        oracleAddress,
		UniqueId:             uniqueID,
		Endpoint:             endpoint,
		OracleCommissionRate: oracleCommissionRate,
	}
}

func (m *Oracle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return sdkerrors.Wrapf(err, "oracle address is invalid. address: %s", m.OracleAddress)
	}
	if len(m.UniqueId) == 0 {
		return fmt.Errorf("uniqueID is empty")
	}
	if len(m.Endpoint) == 0 {
		return fmt.Errorf("endpoint is empty")
	}
	if m.OracleCommissionRate.IsNegative() {
		return fmt.Errorf("oracle commission rate cannot be negative")
	}
	if m.OracleCommissionRate.GT(sdk.OneDec()) {
		return fmt.Errorf("oracle commission rate cannot be greater than 1")
	}
	return nil
}

func NewOracleRegistration(msg *MsgRegisterOracle) *OracleRegistration {
	return &OracleRegistration{
		UniqueId:               msg.UniqueId,
		OracleAddress:          msg.OracleAddress,
		NodePubKey:             msg.NodePubKey,
		NodePubKeyRemoteReport: msg.NodePubKeyRemoteReport,
		TrustedBlockHeight:     msg.TrustedBlockHeight,
		TrustedBlockHash:       msg.TrustedBlockHash,
		Endpoint:               msg.Endpoint,
		OracleCommissionRate:   msg.OracleCommissionRate,
	}
}

func (m *OracleRegistration) ValidateBasic() error {
	if len(m.UniqueId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "uniqueID is empty")
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

	return nil
}
