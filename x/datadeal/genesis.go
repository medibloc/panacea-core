package datadeal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetNextDealNumber(ctx, genState.NextDealNumber)

	for _, deal := range genState.Deals {
		k.SetDeal(ctx, deal)
	}

	for _, dataCert := range genState.DataCerts {
		k.SetDataCert(ctx, dataCert.UnsignedCert.DealId, dataCert)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	deals, err := k.ListDeals(ctx)
	if err != nil {
		panic(err)
	}

	dealsMap := make(map[uint64]types.Deal)
	for _, deal := range deals {
		dealsMap[deal.DealId] = deal
	}

	dataCerts, err := k.ListDataCerts(ctx)
	if err != nil {
		panic(err)
	}

	dataCertMap := make(map[string]types.DataCert)
	for _, dataCert := range dataCerts {
		dataKey := types.GetKeyPrefixDataCert(dataCert.UnsignedCert.DealId, dataCert.UnsignedCert.DataHash)
		dataCertMap[string(dataKey)] = dataCert
	}

	nextDealNum, err := k.GetNextDealNumber(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Deals:          dealsMap,
		DataCerts:      dataCertMap,
		NextDealNumber: nextDealNum,
	}
}
