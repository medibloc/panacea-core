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
	KeyDataPoolDeposit            = []byte("datapooldeposit")
	KeyDataPoolCodeID             = []byte("datapoolcodeid")
	KeyDataPoolNFTContractAddress = []byte("datapoolnftcontractaddress")
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
		DataPoolCodeId:  0,
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
		paramtypes.NewParamSetPair(KeyDataPoolCodeID, &p.DataPoolCodeId, validateDataPoolCodeID),
		paramtypes.NewParamSetPair(KeyDataPoolNFTContractAddress, &p.DataPoolNftContractAddress, validateDataPoolNFTContractAddress),
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

func validateDataPoolCodeID(i interface{}) error {
	codeID, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if codeID < 0 {
		return fmt.Errorf("code id must be non-negative: %d", codeID)
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
