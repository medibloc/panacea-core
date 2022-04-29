package types

import (
	"github.com/medibloc/panacea-core/v2/types/assets"

	"strconv"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	PENDING = "PENDING"
	ACTIVE  = "ACTIVE"
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

// AppendPoolID adds the poolID.
func (d *InstantRevenueDistribute) AppendPoolID(poolID uint64) {
	// Check duplicate
	for _, existPoolID := range d.PoolIds {
		if existPoolID == poolID {
			return
		}
	}
	d.PoolIds = append(d.PoolIds, poolID)
}

func (d *InstantRevenueDistribute) IsEmpty() bool {
	return d.PoolIds == nil || len(d.PoolIds) == 0
}

func (d *InstantRevenueDistribute) RemovePreviousIndex(idx int) {
	if d.IsEmpty() {
		return
	}

	d.PoolIds = d.PoolIds[idx:]
}

func (m *SalesHistory) AddDataHash(dataHash []byte) {
	m.DataHashes = append(m.DataHashes, dataHash)
}

func (m *SalesHistory) DataCount() int {
	return len(m.DataHashes)
}