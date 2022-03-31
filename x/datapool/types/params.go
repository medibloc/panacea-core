package types

import (
	"fmt"

	"github.com/medibloc/panacea-core/v2/types/assets"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultDataPoolDeposit = sdk.Coins{sdk.NewInt64Coin(assets.MicroMedDenom, 10000000000)} // 10000 MED
)

var (
	KeyDataPoolDeposit = []byte("DataPoolDeposit")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(dataPoolDeposit sdk.Coins) Params {
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
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid data pool deposit: %+v", i)
	}

	return nil
}
