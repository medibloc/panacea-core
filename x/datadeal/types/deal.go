package types

import (
	"strconv"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ACTIVE    = "ACTIVE"    // When deal is activated.
	INACTIVE  = "INACTIVE"  // When deal is deactivated.
	COMPLETED = "COMPLETED" // When deal is completed.
)

func NewDeal(dealID uint64, deal Deal) Deal {

	dealAddress := NewDealAddress(dealID)

	return Deal{
		DealId:         dealID,
		DealAddress:    dealAddress.String(),
		DataSchema:     deal.GetDataSchema(),
		Budget:         deal.GetBudget(),
		TrustedOracles: deal.GetTrustedOracles(),
		MaxNumData:     deal.GetMaxNumData(),
		CurNumData:     0,
		Owner:          deal.GetOwner(),
		Status:         ACTIVE,
	}
}

func NewDealAddress(dealID uint64) sdk.AccAddress {
	dealKey := "deal" + strconv.FormatUint(dealID, 10)
	return authtypes.NewModuleAddress(dealKey)
}
