package types

import (
	"fmt"

	"github.com/medibloc/panacea-core/v2/types/assets"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultDataPoolDeposit = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000000000))
)

var (
	KeyDataPoolDeposit = []byte("datapooldeposit")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(dataPoolDeposit sdk.Coin) Params {
	return Params{
		DataPoolDeposit: dataPoolDeposit,
	}
}

func DefaultParams() Params {
	return Params{
		DataPoolDeposit: DefaultDataPoolDeposit,
	}
}

func (p Params) Validate() error {
	if err := validateDataPoolDeposit(p.DataPoolDeposit); err != nil {
		return err
	}

	return nil
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDataPoolDeposit, &p.DataPoolDeposit, validateDataPoolDeposit),
	}
}

func validateDataPoolDeposit(i interface{}) error {
	deposit, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if deposit.Validate() != nil {
		return fmt.Errorf("invalid data pool deposit: %+v", i)
	}

	return nil
}
