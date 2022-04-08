package types

import (
	"fmt"
	"strconv"
	"strings"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/v044_temp/address"
)

const (
	PENDING = "PENDING"
	ACTIVE  = "ACTIVE"
)

func NewPool(poolID uint64, curator sdk.AccAddress, poolParams PoolParams) *Pool {
	poolAddress := NewPoolAddress(poolID)

	return &Pool{
		PoolId:        poolID,
		PoolAddress:   poolAddress.String(),
		Round:         1,
		PoolParams:    &poolParams,
		CurNumData:    0,
		NumIssuedNfts: 0,
		Curator:       curator.String(),
		Status:        PENDING,
	}
}

func NewPoolAddress(poolID uint64) sdk.AccAddress {
	poolKey := "pool" + strconv.FormatUint(poolID, 10)
	return authtypes.NewModuleAddress(poolKey)
}

func AccPoolAddressFromBech32(address string) (sdk.AccAddress, error) {
	if len(strings.TrimSpace(address)) == 0 {
		return sdk.AccAddress{}, fmt.Errorf("empty address string is not allowed")
	}

	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	bz, err := sdk.GetFromBech32(address, bech32PrefixAccAddr)
	if err != nil {
		return nil, err
	}

	err = verifyPoolAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return sdk.AccAddress(bz), nil
}

func verifyPoolAddressFormat(bz []byte) error {
	if len(bz) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, "address cannot be empty")
	}

	if len(bz) > address.MaxAddrLen {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", address.MaxAddrLen, len(bz))
	}

	verifier := sdk.GetConfig().GetAddressVerifier()
	if verifier != nil {
		return verifier(bz)
	}

	return nil
}

func GetModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(ModuleName)
}
