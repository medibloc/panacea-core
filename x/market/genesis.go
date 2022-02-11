package market

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/market/keeper"
	"github.com/medibloc/panacea-core/v2/x/market/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetNextDealNumber(ctx, genState.NextDealNumber)

	for _, deal := range genState.Deals {
		k.SetDeal(ctx, *deal)
	}

	for _, dataCertificate := range genState.DataCertificates {
		k.SetDataCertificate(ctx, dataCertificate.UnsignedCert.DealId, *dataCertificate)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	deals, err := k.ListDeals(ctx)
	if err != nil {
		panic(err)
	}

	dealsMap := make(map[uint64]*types.Deal)
	for _, deal := range deals {
		dealsMap[deal.DealId] = &deal
	}

	dataCertificates, err := k.ListDataCertificates(ctx)
	if err != nil {
		panic(err)
	}

	dataCertificateMap := make(map[string]*types.DataValidationCertificate)
	for _, dataCertificate := range dataCertificates {
		dataKey := types.CombineKeys(sdk.Uint64ToBigEndian(dataCertificate.UnsignedCert.DealId), dataCertificate.UnsignedCert.DataHash)
		dataCertificateMap[string(dataKey)] = &dataCertificate
	}

	return &types.GenesisState{
		Deals:            dealsMap,
		DataCertificates: dataCertificateMap,
		NextDealNumber:   k.GetNextDealNumber(ctx),
	}
}
