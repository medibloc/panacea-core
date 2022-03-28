package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/datadeal/v044_temp/address"
)

const (
	PENDING = "PENDING"
	ACTIVE  = "ACTIVE"
)

func NewPool(poolID uint64, curator sdk.AccAddress, poolParams PoolParams) *Pool {
	poolAddress := newPoolAddress(poolID)

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

func newPoolAddress(poolID uint64) sdk.AccAddress {
	key := append([]byte("pool"), sdk.Uint64ToBigEndian(poolID)...)
	return address.Module(ModuleName, key)
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
	verifier := sdk.GetConfig().GetAddressVerifier()
	if verifier != nil {
		return verifier(bz)
	}

	if len(bz) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, "address cannot be empty")
	}

	if len(bz) > address.MaxAddrLen {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", address.MaxAddrLen, len(bz))
	}

	return nil
}

func GetModuleAddress() sdk.AccAddress {
	// 20 byte length address
	return address.Module(ModuleName, []byte("module-account"))[:sdk.AddrLen]
}
