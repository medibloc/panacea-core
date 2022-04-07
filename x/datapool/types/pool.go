package types

import (
	"fmt"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	PENDING          = "PENDING"
	ACTIVE           = "ACTIVE"
	ShareTokenPrefix = "datapool"
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

func GetModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(ModuleName)
}

func GetAccumPoolShareToken(poolID, amount uint64) sdk.Coin {
	return sdk.NewCoin(fmt.Sprintf(ShareTokenPrefix+"/%v", poolID), sdk.NewIntFromUint64(amount))
}
