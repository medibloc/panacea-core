package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyDealDeactivationParam = []byte("DealDeactivationParam")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() Params {
	return Params{
		DealDeactivationParam: 3,
	}
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDealDeactivationParam, &p.DealDeactivationParam, validateDealDeactivationParam),
	}
}

func (p Params) Validate() error {
	if err := validateDealDeactivationParam(p.DealDeactivationParam); err != nil {
		return err
	}

	return nil
}

func validateDealDeactivationParam(i interface{}) error {
	dealDeactivationParam, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if dealDeactivationParam < 0 {
		return fmt.Errorf("deal deactivation param cannot be negative")
	}
	return nil
}
