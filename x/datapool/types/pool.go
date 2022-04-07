package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	PENDING = "PENDING"
	ACTIVE  = "ACTIVE"

	ShareTokenPrefix = "DP"
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

func GetDenomOfShareToken(poolID uint64) string {
	return fmt.Sprintf(ShareTokenPrefix+"/%v", poolID)
}

func GetAccumPoolShareToken(poolID, amount uint64) sdk.Coin {
	return sdk.NewCoin(GetDenomOfShareToken(poolID), sdk.NewIntFromUint64(amount))
}

func (d *DelayedRevenueDistribute) AppendPoolID(poolID uint64) {
	d.PoolIds = append(d.PoolIds, poolID)
}

func (d *DelayedRevenueDistribute) IsEmpty() bool {
	return d.PoolIds == nil || len(d.PoolIds) == 0
}

func (d *DelayedRevenueDistribute) RemovePreviousIndex(idx int) {
	if d.IsEmpty() {
		return
	}

	d.PoolIds = d.PoolIds[idx:]
}
