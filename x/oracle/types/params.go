package types

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyOraclePublicKey          = []byte("OraclePublicKey")
	KeyOraclePubKeyRemoteReport = []byte("OraclePubKeyRemoteReport")
	KeyUniqueID                 = []byte("UniqueID")
	KeyVoteParams               = []byte("VoteParams")
	KeySlashParams              = []byte("SlashParams")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() Params {
	return Params{
		OraclePublicKey:          nil,
		OraclePubKeyRemoteReport: nil,
		UniqueId:                 "",
		VoteParams: VoteParams{
			VotingPeriod: 30 * time.Second,
			JailPeriod:   10 * time.Minute,
			Threshold:    sdk.NewDec(2).Quo(sdk.NewDec(3)),
		},
		SlashParams: SlashParams{
			SlashFractionDowntime: sdk.NewDecWithPrec(2, 1),
			SlashFractionForgery:  sdk.NewDecWithPrec(1, 1),
		},
	}
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyOraclePublicKey, &p.OraclePublicKey, validateOraclePublicKey),
		paramtypes.NewParamSetPair(KeyOraclePubKeyRemoteReport, &p.OraclePubKeyRemoteReport, validateOraclePubKeyRemoteReport),
		paramtypes.NewParamSetPair(KeyUniqueID, &p.UniqueId, validateUniqueID),
		paramtypes.NewParamSetPair(KeyVoteParams, &p.VoteParams, validateVoteParams),
		paramtypes.NewParamSetPair(KeySlashParams, &p.SlashParams, validateSlashParams),
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
	if err := validateVoteParams(p.VoteParams); err != nil {
		return err
	}
	if err := validateSlashParams(p.SlashParams); err != nil {
		return err
	}

	return nil
}

func validateOraclePublicKey(i interface{}) error {
	oraclePubKey, ok := i.([]byte)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if oraclePubKey != nil {
		_, err := btcec.ParsePubKey(oraclePubKey, btcec.S256())
		if err != nil {
			return err
		}
	}

	return nil
}

func validateOraclePubKeyRemoteReport(i interface{}) error {
	_, ok := i.([]byte)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
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

func validateVoteParams(i interface{}) error {
	voteParams, ok := i.(VoteParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if voteParams.VotingPeriod == 0 {
		return fmt.Errorf("'votingPeriod' cannot be set to zero")
	}

	if voteParams.JailPeriod == 0 {
		return fmt.Errorf("'jailPeriod' cannot be set to zero")
	}

	if sdk.NewDec(0).Equal(voteParams.Threshold) {
		return fmt.Errorf("'threshold' cannot be set to zero")
	}

	return nil
}

func validateSlashParams(i interface{}) error {
	slashParams, ok := i.(SlashParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if slashParams.SlashFractionDowntime.IsNegative() {
		return fmt.Errorf("'slashFactionDowntime' cannot be negative: %s", slashParams.SlashFractionDowntime)
	} else if slashParams.SlashFractionDowntime.GT(sdk.OneDec()) {
		return fmt.Errorf("'slashFactionDowntime' rate cannot be greater than 100%%: %s", slashParams.SlashFractionDowntime)
	}

	if slashParams.SlashFractionForgery.IsNegative() {
		return fmt.Errorf("'slashFractionForgery' cannot be negative: %s", slashParams.SlashFractionForgery)
	} else if slashParams.SlashFractionForgery.GT(sdk.OneDec()) {
		return fmt.Errorf("'slashFractionForgery' rate cannot be greater than 100%%: %s", slashParams.SlashFractionForgery)
	}

	return nil
}
