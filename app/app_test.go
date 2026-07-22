package app_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govmodule "github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	"github.com/medibloc/panacea-core/v2/app/upgrades/v2_3_0"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	pnftlegacy "github.com/medibloc/panacea-core/v2/x/pnft/legacy"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

func TestRuntimeModulesHaveBootstrapBasics(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	runtimeBasics := testApp.BasicManager()
	moduleNames := make([]string, 0, len(testApp.ModuleManager.Modules))
	for name := range testApp.ModuleManager.Modules {
		moduleNames = append(moduleNames, name)
		require.Contains(t, runtimeBasics, name, "runtime module %q has no derived basic", name)
		require.Contains(t, panaceaapp.ModuleBasics, name, "runtime module %q is missing from bootstrap basics", name)
	}

	runtimeBasicNames := make([]string, 0, len(runtimeBasics))
	for name := range runtimeBasics {
		runtimeBasicNames = append(runtimeBasicNames, name)
	}
	require.ElementsMatch(t, moduleNames, runtimeBasicNames)

	var bootstrapOnly []string
	for name := range panaceaapp.ModuleBasics {
		if _, ok := runtimeBasics[name]; !ok {
			bootstrapOnly = append(bootstrapOnly, name)
		}
	}
	require.ElementsMatch(t, []string{ibctm.ModuleName}, bootstrapOnly)
}

func TestCapabilityWiring(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	require.True(t, testApp.CapabilityKeeper.HasModule(ibcexported.ModuleName))
	require.True(t, testApp.CapabilityKeeper.HasModule(ibctransfertypes.ModuleName))
	require.True(t, testApp.CapabilityKeeper.IsSealed())
	require.Contains(t, testApp.ModuleManager.Modules, capabilitytypes.ModuleName)

	require.NotEmpty(t, testApp.ModuleManager.OrderBeginBlockers)
	require.Equal(t, capabilitytypes.ModuleName, testApp.ModuleManager.OrderBeginBlockers[0])
	require.NotEmpty(t, testApp.ModuleManager.OrderInitGenesis)
	require.Equal(t, capabilitytypes.ModuleName, testApp.ModuleManager.OrderInitGenesis[0])
	require.NotEmpty(t, testApp.ModuleManager.OrderExportGenesis)
	require.Equal(t, capabilitytypes.ModuleName, testApp.ModuleManager.OrderExportGenesis[0])
}

func TestNFTStoreAndPNFTCompatibilityWiring(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	require.NotContains(t, testApp.GetKVStoreKey(), pnfttypes.StoreKey)
	require.Contains(t, testApp.GetKVStoreKey(), nfttypes.StoreKey)
	require.Contains(t, testApp.GetKVStoreKey(), nfttypes.PolicyStoreKey)
	require.NotContains(t, testApp.ModuleManager.Modules, pnfttypes.ModuleName)
	require.NotContains(t, testApp.ModuleManager.Modules, nfttypes.ModuleName)
	require.NotContains(t, testApp.BasicManager(), pnfttypes.ModuleName)
	require.NotContains(t, testApp.BasicManager(), nfttypes.ModuleName)
	require.NotContains(t, panaceaapp.ModuleBasics, pnfttypes.ModuleName)
	require.NotContains(t, panaceaapp.ModuleBasics, nfttypes.ModuleName)
	require.NotContains(t, testApp.DefaultGenesis(), pnfttypes.ModuleName)
	require.NotContains(t, testApp.ModuleManager.GetVersionMap(), pnfttypes.ModuleName)
	require.NotContains(t, testApp.ModuleManager.OrderInitGenesis, pnfttypes.ModuleName)
	require.NotContains(t, testApp.ModuleManager.OrderExportGenesis, pnfttypes.ModuleName)

	require.NoError(t, testApp.LoadLatestVersion())
	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Time: time.Now()})
	exported, err := testApp.NFTKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exported.NftState)
	require.Empty(t, exported.ClassPolicies)
}

func TestPNFTMsgRouteUsesLegacyRejectionServer(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	msg := &pnfttypes.MsgMintPNFTRequest{
		DenomId: "legacy-denom",
		Id:      "legacy-pnft",
		Name:    "Legacy PNFT",
		Creator: sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String(),
	}
	handler := testApp.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)

	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Time: time.Now()})
	response, err := handler(ctx, msg)
	require.Nil(t, response)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, pnftlegacy.DisabledErrorMessage)
}

func TestLegacyPNFTProposalExecutionFails(t *testing.T) {
	t.Run("governance proposal", func(t *testing.T) {
		testApp, voter, ctx := newPNFTProposalTestApp(t)
		legacyMsg := newLegacyProposalPNFTMsg(authtypes.NewModuleAddress(govtypes.ModuleName).String())

		proposal, err := testApp.GovKeeper.SubmitProposal(
			ctx,
			[]sdk.Msg{legacyMsg},
			"",
			"Legacy PNFT proposal",
			"Execute a legacy PNFT message",
			voter,
			false,
		)
		require.NoError(t, err)
		require.NoError(t, testApp.GovKeeper.ActivateVotingPeriod(ctx, proposal))
		require.NoError(t, testApp.GovKeeper.AddVote(
			ctx,
			proposal.Id,
			voter,
			govv1.NewNonSplitVoteOption(govv1.OptionYes),
			"",
		))

		stored, err := testApp.GovKeeper.Proposals.Get(ctx, proposal.Id)
		require.NoError(t, err)
		require.NotNil(t, stored.VotingEndTime)

		executionCtx := ctx.WithBlockTime(stored.VotingEndTime.Add(time.Nanosecond))
		require.NoError(t, govmodule.EndBlocker(executionCtx, &testApp.GovKeeper))

		failed, err := testApp.GovKeeper.Proposals.Get(executionCtx, proposal.Id)
		require.NoError(t, err)
		require.Equal(t, govv1.StatusFailed, failed.Status)
		require.Contains(t, failed.FailedReason, pnftlegacy.DisabledErrorMessage)
	})

	t.Run("group proposal", func(t *testing.T) {
		testApp, member, ctx := newPNFTProposalTestApp(t)

		createGroup := &group.MsgCreateGroupWithPolicy{
			Admin: member.String(),
			Members: []group.MemberRequest{{
				Address: member.String(),
				Weight:  "1",
			}},
		}
		require.NoError(t, createGroup.SetDecisionPolicy(
			group.NewThresholdDecisionPolicy("1", time.Hour, 0),
		))
		created, err := testApp.GroupKeeper.CreateGroupWithPolicy(ctx, createGroup)
		require.NoError(t, err)

		legacyMsg := newLegacyProposalPNFTMsg(created.GroupPolicyAddress)
		submit, err := group.NewMsgSubmitProposal(
			created.GroupPolicyAddress,
			[]string{member.String()},
			[]sdk.Msg{legacyMsg},
			"",
			group.Exec_EXEC_UNSPECIFIED,
			"Legacy PNFT proposal",
			"Execute a legacy PNFT message",
		)
		require.NoError(t, err)
		submitted, err := testApp.GroupKeeper.SubmitProposal(ctx, submit)
		require.NoError(t, err)
		_, err = testApp.GroupKeeper.Vote(ctx, &group.MsgVote{
			ProposalId: submitted.ProposalId,
			Voter:      member.String(),
			Option:     group.VOTE_OPTION_YES,
		})
		require.NoError(t, err)

		executionCtx := ctx.WithEventManager(sdk.NewEventManager())
		executed, err := testApp.GroupKeeper.Exec(executionCtx, &group.MsgExec{
			ProposalId: submitted.ProposalId,
			Executor:   member.String(),
		})
		require.NoError(t, err)
		require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_FAILURE, executed.Result)

		stored, err := testApp.GroupKeeper.Proposal(executionCtx, &group.QueryProposalRequest{
			ProposalId: submitted.ProposalId,
		})
		require.NoError(t, err)
		require.Equal(t, group.PROPOSAL_STATUS_ACCEPTED, stored.Proposal.Status)
		require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_FAILURE, stored.Proposal.ExecutorResult)

		var executionEvent *group.EventExec
		for _, event := range executionCtx.EventManager().ABCIEvents() {
			parsed, parseErr := sdk.ParseTypedEvent(event)
			require.NoError(t, parseErr)
			if typed, ok := parsed.(*group.EventExec); ok {
				executionEvent = typed
				break
			}
		}
		require.NotNil(t, executionEvent)
		require.Contains(t, executionEvent.Logs, pnftlegacy.DisabledErrorMessage)
	})
}

func newPNFTProposalTestApp(t *testing.T) (*panaceaapp.App, sdk.AccAddress, sdk.Context) {
	t.Helper()

	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	validatorSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)
	privateKey := secp256k1.GenPrivKey()
	account := authtypes.NewBaseAccount(privateKey.PubKey().Address().Bytes(), privateKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: account.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000)),
	}
	genesisState, err := simtestutil.GenesisStateWithValSet(
		testApp.AppCodec(),
		testApp.DefaultGenesis(),
		validatorSet,
		[]authtypes.GenesisAccount{account},
		balance,
	)
	require.NoError(t, err)
	genesisBytes, err := cmtjson.Marshal(genesisState)
	require.NoError(t, err)

	blockTime := time.Now().UTC()
	_, err = testApp.InitChain(&abci.RequestInitChain{
		Time:            blockTime,
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   genesisBytes,
	})
	require.NoError(t, err)
	_, err = testApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             1,
		Time:               blockTime,
		NextValidatorsHash: validatorSet.Hash(),
	})
	require.NoError(t, err)
	_, err = testApp.Commit()
	require.NoError(t, err)

	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Height: 2, Time: blockTime})
	return testApp, account.GetAddress(), ctx
}

func newLegacyProposalPNFTMsg(authority string) *pnfttypes.MsgMintPNFTRequest {
	return &pnfttypes.MsgMintPNFTRequest{
		DenomId: "legacy-denom",
		Id:      "legacy-pnft",
		Name:    "Legacy PNFT",
		Creator: authority,
	}
}

func TestPNFTQueryRoutesAreNotRegistered(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	for _, path := range []string{
		"/panacea.pnft.v2.Query/Denoms",
		"/panacea.pnft.v2.Query/DenomsByOwner",
		"/panacea.pnft.v2.Query/Denom",
		"/panacea.pnft.v2.Query/PNFTs",
		"/panacea.pnft.v2.Query/PNFTsByDenomOwner",
		"/panacea.pnft.v2.Query/PNFT",
	} {
		t.Run(path, func(t *testing.T) {
			require.Nil(t, testApp.GRPCQueryRouter().Route(path))
		})
	}
}

func TestFeeGrantWiringWithStandaloneModule(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	require.Contains(t, testApp.GetKVStoreKey(), feegrant.StoreKey)
	require.Contains(t, testApp.ModuleManager.Modules, feegrant.ModuleName)

	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Time: time.Now()})
	granter := sdk.AccAddress(bytes.Repeat([]byte{3}, 20))
	grantee := sdk.AccAddress(bytes.Repeat([]byte{4}, 20))
	spendLimit := sdk.NewCoins(sdk.NewInt64Coin("umed", 100))

	msg, err := feegrant.NewMsgGrantAllowance(
		&feegrant.BasicAllowance{SpendLimit: spendLimit},
		granter,
		grantee,
	)
	require.NoError(t, err)

	handler := testApp.MsgServiceRouter().Handler(msg)
	require.NotNil(t, handler)
	_, err = handler(ctx, msg)
	require.NoError(t, err)

	allowance, err := testApp.FeeGrantKeeper.GetAllowance(ctx, granter, grantee)
	require.NoError(t, err)
	basicAllowance, ok := allowance.(*feegrant.BasicAllowance)
	require.True(t, ok)
	require.Equal(t, spendLimit, basicAllowance.SpendLimit)
}

func TestUpgradeWiringWithStandaloneModule(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	require.Contains(t, testApp.GetKVStoreKey(), upgradetypes.StoreKey)
	require.Contains(t, testApp.ModuleManager.Modules, upgradetypes.ModuleName)
	require.Equal(t, []string{upgradetypes.ModuleName}, testApp.ModuleManager.OrderPreBlockers)
	require.Same(t, testApp.BaseApp, testApp.UpgradeKeeper.GetVersionSetter())

	const (
		upgradeName   = "test-standalone-upgrade"
		upgradeHeight = int64(10)
	)
	blockTime := time.Now()
	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Height: upgradeHeight, Time: blockTime}).
		WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})
	require.NoError(t, testApp.UpgradeKeeper.SetModuleVersionMap(ctx, testApp.ModuleManager.GetVersionMap()))

	handlerCalled := false
	testApp.UpgradeKeeper.SetUpgradeHandler(
		upgradeName,
		func(_ context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			handlerCalled = true
			return fromVM, nil
		},
	)

	plan := upgradetypes.Plan{Name: upgradeName, Height: upgradeHeight}
	require.NoError(t, testApp.UpgradeKeeper.ScheduleUpgrade(ctx, plan))
	storedPlan, err := testApp.UpgradeKeeper.GetUpgradePlan(ctx)
	require.NoError(t, err)
	require.Equal(t, plan, storedPlan)

	response, err := testApp.PreBlocker(ctx, &abci.RequestFinalizeBlock{})
	require.NoError(t, err)
	require.True(t, response.ConsensusParamsChanged)
	require.True(t, handlerCalled)

	doneHeight, err := testApp.UpgradeKeeper.GetDoneHeight(ctx, upgradeName)
	require.NoError(t, err)
	require.Equal(t, upgradeHeight, doneHeight)
	_, err = testApp.UpgradeKeeper.GetUpgradePlan(ctx)
	require.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)
}

func TestV230UpgradeDropsLegacyPNFTModuleVersion(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)
	require.NoError(t, testApp.LoadLatestVersion())

	const upgradeHeight = int64(10)
	blockTime := time.Now()
	ctx := testApp.NewUncachedContext(false, cmtproto.Header{Height: upgradeHeight, Time: blockTime}).
		WithHeaderInfo(header.Info{Height: upgradeHeight, Time: blockTime})

	fromVM := testApp.ModuleManager.GetVersionMap()
	fromVM[pnfttypes.ModuleName] = 1
	require.NoError(t, testApp.UpgradeKeeper.SetModuleVersionMap(ctx, fromVM))
	require.NoError(t, testApp.UpgradeKeeper.ScheduleUpgrade(ctx, upgradetypes.Plan{
		Name:   v2_3_0.UpgradeName,
		Height: upgradeHeight,
	}))

	_, err := testApp.PreBlocker(ctx, &abci.RequestFinalizeBlock{})
	require.NoError(t, err)

	toVM, err := testApp.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.NotContains(t, toVM, pnfttypes.ModuleName)
}
