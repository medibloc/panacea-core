package types

import (
	"encoding/base64"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyOraclePublicKey          = []byte("OraclePublicKey")
	KeyOraclePubKeyRemoteReport = []byte("OraclePubKeyRemoteReport")
	KeyUniqueID                 = []byte("UniqueID")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() Params {
	return Params{
		OraclePublicKey:          "",
		OraclePubKeyRemoteReport: "",
		UniqueId:                 "",
	}
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyOraclePublicKey, &p.OraclePublicKey, validateOraclePublicKey),
		paramtypes.NewParamSetPair(KeyOraclePubKeyRemoteReport, &p.OraclePubKeyRemoteReport, validateOraclePubKeyRemoteReport),
		paramtypes.NewParamSetPair(KeyUniqueID, &p.UniqueId, validateUniqueID),
	}
}

func (p Params) Validate() error {
	if err := validateOraclePublicKey(p.OraclePublicKey); err != nil {
		return err
	}
	if err := validateOraclePubKeyRemoteReport(p.OraclePubKeyRemoteReport); err != nil {
		return err
	}
	if err := validateUniqueID(p.UniqueId); err != nil {
		return err
	}

	return nil
}

func validateOraclePublicKey(i interface{}) error {
	pubKeyBase64, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if pubKeyBase64 != "" {
		oraclePubKeyBz, err := base64.StdEncoding.DecodeString(pubKeyBase64)
		if err != nil {
			return err
		}

		if _, err := btcec.ParsePubKey(oraclePubKeyBz, btcec.S256()); err != nil {
			return err
		}
	}

	return nil
}

func validateOraclePubKeyRemoteReport(i interface{}) error {
	reportBase64, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if reportBase64 != "" {
		if _, err := base64.StdEncoding.DecodeString(reportBase64); err != nil {
			return err
		}
	}

	return nil
}

func validateUniqueID(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
