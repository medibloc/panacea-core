package types

import (
	"github.com/medibloc/panacea-core/v2/types/assets"

	"fmt"
	"strconv"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	PENDING = "PENDING"
	ACTIVE  = "ACTIVE"

	ShareTokenPrefix = "DP"
)

var ZeroFund = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0))

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

func NewDataPassRedeemReceipt(poolID, round, nftID uint64, redeemer string) *DataPassRedeemReceipt {
	return &DataPassRedeemReceipt{
		PoolId:   poolID,
		Round:    round,
		NftId:    nftID,
		Redeemer: redeemer,
	}
}
