package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultDataPoolCommissionRate = sdk.NewDecWithPrec(1, 2) // default 1%
)

var (
	KeyDataPoolCommissionRate     = []byte("DataPoolCommissionRate")
	KeyDataPoolCodeID             = []byte("DataPoolCodeId")
	KeyDataPoolNFTContractAddress = []byte("DataPoolNftContractAddress")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(dataPoolCuratorCommissionRate sdk.Dec) Params {
	return Params{
		DataPoolCommissionRate: dataPoolCuratorCommissionRate,
	}
}

func DefaultParams() Params {
	return Params{
		DataPoolCodeId:         0,
		DataPoolCommissionRate: DefaultDataPoolCommissionRate,
	}
}

func (p Params) Validate() error {
	if err := validateDataPoolCommissionRate(p.DataPoolCommissionRate); err != nil {
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
		paramtypes.NewParamSetPair(KeyDataPoolCommissionRate, &p.DataPoolCommissionRate, validateDataPoolCommissionRate),
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

func validateDataPoolCommissionRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("commission rate cannot be negative: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("commission rate cannot be greater than 100%%: %s", v)
	}

	return nil
}
