package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultDepositRate = sdk.NewDecWithPrec(1, 1) // default 10%
)

var (
	KeyDepositRate                = []byte("DepositRate")
	KeyDataPoolCodeID             = []byte("DataPoolCodeId")
	KeyDataPoolNFTContractAddress = []byte("DataPoolNftContractAddress")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(depositRate sdk.Dec) Params {
	return Params{
		DepositRate: depositRate,
	}
}

func DefaultParams() Params {
	return Params{
		DepositRate:    DefaultDepositRate,
		DataPoolCodeId: 0,
	}
}

func (p Params) Validate() error {
	if err := validateDepositRate(p.DepositRate); err != nil {
		return err
	}

	if err := validateDataPoolCodeID(p.DataPoolCodeId); err != nil {
		return err
	}

	if err := validateDataPoolNFTContractAddress(p.DataPoolNftContractAddress); err != nil {
		return err
	}

	return nil
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDepositRate, &p.DepositRate, validateDepositRate),
		paramtypes.NewParamSetPair(KeyDataPoolCodeID, &p.DataPoolCodeId, validateDataPoolCodeID),
		paramtypes.NewParamSetPair(KeyDataPoolNFTContractAddress, &p.DataPoolNftContractAddress, validateDataPoolNFTContractAddress),
	}
}

func validateDataPoolCodeID(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDataPoolNFTContractAddress(i interface{}) error {
	addr, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if addr != "" {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return fmt.Errorf("invalid NFT contract address: %s", addr)
		}
	}

	return nil
}

func validateDepositRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("deposit rate cannot be negative: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("deposit rate cannot be greater than 100%%: %s", v)
	}

	return nil
}
