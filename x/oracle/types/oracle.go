package types

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewOracle(oracleAddress, uniqueID, endpoint string, rate, maxRate, maxChangeRate sdk.Dec, updatedAt time.Time) *Oracle {
	return &Oracle{
		OracleAddress:                 oracleAddress,
		UniqueId:                      uniqueID,
		Endpoint:                      endpoint,
		UpdateTime:                    updatedAt,
		OracleCommissionRate:          rate,
		OracleCommissionMaxRate:       maxRate,
		OracleCommissionMaxChangeRate: maxChangeRate,
	}
}

func (m *Oracle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return sdkerrors.Wrapf(err, "oracle address is invalid. address: %s", m.OracleAddress)
	}
	if len(m.UniqueId) == 0 {
		return fmt.Errorf("uniqueID is empty")
	}

	if m.OracleCommissionMaxRate.LT(sdk.ZeroDec()) || m.OracleCommissionMaxRate.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "OracleCommissionMaxRate must be between 0 and 1")
	}

	if m.OracleCommissionRate.LT(sdk.ZeroDec()) || m.OracleCommissionRate.GT(m.OracleCommissionMaxRate) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleCommissionRate must be between 0 and OracleCommissionMaxRate")
	}

	if m.OracleCommissionMaxChangeRate.LT(sdk.ZeroDec()) || m.OracleCommissionMaxChangeRate.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "OracleCommissionMaxChangeRate must be between 0 and 1")
	}
	return nil
}

func NewOracleRegistration(msg *MsgRegisterOracle) *OracleRegistration {
	return &OracleRegistration{
		UniqueId:                      msg.UniqueId,
		OracleAddress:                 msg.OracleAddress,
		NodePubKey:                    msg.NodePubKey,
		NodePubKeyRemoteReport:        msg.NodePubKeyRemoteReport,
		TrustedBlockHeight:            msg.TrustedBlockHeight,
		TrustedBlockHash:              msg.TrustedBlockHash,
		Endpoint:                      msg.Endpoint,
		OracleCommissionRate:          msg.OracleCommissionRate,
		OracleCommissionMaxRate:       msg.OracleCommissionMaxRate,
		OracleCommissionMaxChangeRate: msg.OracleCommissionMaxChangeRate,
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

	if m.OracleCommissionMaxRate.LT(sdk.ZeroDec()) || m.OracleCommissionMaxRate.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "OracleCommissionMaxRate must be between 0 and 1")
	}

	if m.OracleCommissionRate.LT(sdk.ZeroDec()) || m.OracleCommissionRate.GT(m.OracleCommissionMaxRate) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracleCommissionRate must be between 0 and OracleCommissionMaxRate")
	}

	if m.OracleCommissionMaxChangeRate.LT(sdk.ZeroDec()) || m.OracleCommissionMaxChangeRate.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "OracleCommissionMaxChangeRate must be between 0 and 1")
	}

	return nil
}

// ValidateOracleCommission validate an oracle's commission rate.
// An error is returned if the new commission rate is invalid.
func (m *Oracle) ValidateOracleCommission(blockTime time.Time, newRate sdk.Dec) error {
	switch {
	case blockTime.Sub(m.UpdateTime).Hours() < 24:
		return ErrCommissionUpdateTime

	case newRate.IsNegative():
		// new rate cannot be negative
		return ErrCommissionNegative

	case newRate.GT(m.OracleCommissionMaxRate):
		// new rate cannot be greater than the max rate
		return ErrCommissionGTMaxRate

	case newRate.Sub(m.OracleCommissionRate).GT(m.OracleCommissionMaxChangeRate):
		// new rate % points change cannot be greater than the max change rate
		return ErrCommissionGTMaxChangeRate
	}

	return nil
}

func (m *OracleUpgradeInfo) ShouldExecute(ctx sdk.Context) bool {
	if m.Height > 0 {
		return m.Height == ctx.BlockHeight()
	}
	return false
}
