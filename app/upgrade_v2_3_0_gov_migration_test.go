package app_test

import (
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	"github.com/medibloc/panacea-core/v2/app/upgrades/v2_3_0"
	"github.com/medibloc/panacea-core/v2/types/assets"
)

func TestV230UpgradeConfiguresExpeditedGovernanceParamsForShortLegacyVotingPeriod(t *testing.T) {
	testV230UpgradeGovernanceParams(t, 20*time.Second, 10*time.Second)
}

func TestV230UpgradePreservesValidMigratedGovernanceVotingPeriods(t *testing.T) {
	testV230UpgradeGovernanceParams(t, 3*24*time.Hour, govv1.DefaultExpeditedPeriod)
}

func testV230UpgradeGovernanceParams(t *testing.T, legacyVotingPeriod, expectedExpeditedVotingPeriod time.Duration) {
	t.Helper()

	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	const upgradeHeight = int64(10)
	blockTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := testApp.NewUncachedContext(
		false,
		cmtproto.Header{Height: upgradeHeight, Time: blockTime},
	).WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})

	fromVM := testApp.ModuleManager.GetVersionMap()
	fromVM[govtypes.ModuleName] = 4
	require.NoError(t, testApp.UpgradeKeeper.SetModuleVersionMap(ctx, fromVM))

	expectedMinDeposit := sdk.NewCoins(sdk.NewInt64Coin(assets.MicroMedDenom, 100_000_000_000))
	legacyParams := govv1.DefaultParams()
	legacyParams.VotingPeriod = &legacyVotingPeriod
	legacyParams.MinDeposit = expectedMinDeposit
	legacyParams.ExpeditedMinDeposit = nil
	legacyParams.ExpeditedVotingPeriod = nil
	legacyParams.ExpeditedThreshold = ""
	legacyParams.ProposalCancelRatio = ""
	legacyParams.ProposalCancelDest = ""
	legacyParams.MinDepositRatio = ""
	require.NoError(t, testApp.GovKeeper.Params.Set(ctx, legacyParams))

	require.NoError(t, testApp.UpgradeKeeper.ScheduleUpgrade(ctx, upgradetypes.Plan{
		Name:   v2_3_0.UpgradeName,
		Height: upgradeHeight,
	}))
	_, err := testApp.PreBlocker(ctx, &abci.RequestFinalizeBlock{Height: upgradeHeight})
	require.NoError(t, err)

	toVM, err := testApp.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(5), toVM[govtypes.ModuleName])

	params, err := testApp.GovKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedMinDeposit, sdk.Coins(params.MinDeposit))
	require.Equal(
		t,
		expectedMinDeposit.MulInt(math.NewInt(govv1.DefaultMinExpeditedDepositTokensRatio)),
		sdk.Coins(params.ExpeditedMinDeposit),
	)
	require.NotNil(t, params.VotingPeriod)
	require.Equal(t, legacyVotingPeriod, *params.VotingPeriod)
	require.NotNil(t, params.ExpeditedVotingPeriod)
	require.Equal(t, expectedExpeditedVotingPeriod, *params.ExpeditedVotingPeriod)
	require.NoError(t, params.ValidateBasic())
}
