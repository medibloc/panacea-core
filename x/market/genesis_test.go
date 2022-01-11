package market_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	"github.com/medibloc/panacea-core/v2/x/market"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

var acc1 = secp256k1.GenPrivKey().PubKey().Address()

func TestMarketInitGenesis(t *testing.T) {
	app := panaceaapp.SetUp(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	newDeal, err := makeTestDeal()
	require.NoError(t, err)

	market.InitGenesis(ctx, app.MarketKeeper, types.GenesisState{
		Deals: map[uint64]*types.Deal{
			newDeal.GetDealId(): &newDeal,
		},
		NextDealNumber: 2,
	})

	require.Equal(t, app.MarketKeeper.GetNextDealNumberAndIncrement(ctx), uint64(2))
	require.NoError(t, err)

	dealStored, err := app.MarketKeeper.GetDeal(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, newDeal.GetDealId(), dealStored.GetDealId())
	require.Equal(t, newDeal.GetDealAddress(), dealStored.GetDealAddress())
	require.Equal(t, newDeal.GetBudget(), dealStored.GetBudget())
	require.Equal(t, newDeal.GetWantDataCount(), dealStored.GetWantDataCount())
	require.Equal(t, newDeal.GetOwner(), dealStored.GetOwner())
	require.Equal(t, newDeal.GetStatus(), dealStored.GetStatus())

	_, err = app.MarketKeeper.GetDeal(ctx, 2)
	require.Error(t, err)
}

func makeTestDeal() (types.Deal, error) {
	return types.Deal{
		DealId:               1,
		DealAddress:          types.NewDealAddress(1).String(),
		DataSchema:           nil,
		Budget:               &sdk.Coin{Denom: "umed", Amount: sdk.NewInt(10000000)},
		TrustedDataValidator: nil,
		WantDataCount:        10000,
		CompleteDataCount:    0,
		Owner:                acc1.String(),
		Status:               "ACTIVE",
	}, nil
}
