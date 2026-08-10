package keeper

import (
	"testing"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestExportGenesisRejectsPolicyKeyValueMismatch(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)
	key := collections.Join(classID, "nft-1")
	lifecycle, err := fixture.keeper.lifecycles.Get(fixture.ctx, key)
	require.NoError(t, err)
	lifecycle.NftId = "other"
	require.NoError(t, fixture.keeper.lifecycles.Set(fixture.ctx, key, lifecycle))

	_, err = fixture.keeper.ExportGenesis(fixture.ctx)
	require.ErrorContains(t, err, "lifecycle key does not match value")
}

func TestInitGenesisRejectsModuleAccountsBeforeWriting(t *testing.T) {
	source := newKeeperFixture(t, true, true)
	_, _, _, _, _ = createNFTForBurnTest(t, &source)
	exported, err := source.keeper.ExportGenesis(source.ctx)
	require.NoError(t, err)
	target := newKeeperFixture(t, true, true)
	moduleController := target.accountAddress(t, target.accountKeeper.moduleAddress)
	exported.ClassPolicies[0].Controller = moduleController

	err = target.keeper.InitGenesis(target.ctx, exported)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "must not be a module account")
	require.Empty(t, target.keeper.nftKeeper.GetClasses(target.ctx))
	require.Empty(t, target.ctx.EventManager().Events())
}

func TestInitGenesisRollsBackStandardStateWhenPolicyWriteFails(t *testing.T) {
	for _, test := range []struct {
		name   string
		failAt int
	}{
		{name: "class policy write", failAt: 1},
		{name: "owner class count write", failAt: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := newKeeperFixture(t, true, true)
			classID, _, _, ownerAddress, _ := createNFTForBurnTest(t, &source)
			exported, err := source.keeper.ExportGenesis(source.ctx)
			require.NoError(t, err)
			target := newKeeperFixture(t, true, true)
			setCalls := 0
			failingKeeper := NewKeeper(
				target.cdc,
				runtime.NewKVStoreService(target.nftService),
				failingNthSetStoreService{
					delegate: runtime.NewKVStoreService(target.policyService),
					calls:    &setCalls,
					failAt:   test.failAt,
				},
				target.accountKeeper,
				testBankKeeper{},
				target.moduleAccountAddresses,
			)

			err = failingKeeper.InitGenesis(target.ctx, exported)
			require.ErrorContains(t, err, "forced set failure")
			require.Empty(t, target.keeper.nftKeeper.GetClasses(target.ctx))
			_, found := target.keeper.nftKeeper.GetNFT(target.ctx, classID, "nft-1")
			require.False(t, found)
			require.Empty(t, target.keeper.nftKeeper.GetOwner(target.ctx, classID, "nft-1"))
			require.Zero(t, target.keeper.nftKeeper.GetTotalSupply(target.ctx, classID))
			require.Zero(t, target.keeper.nftKeeper.GetBalance(target.ctx, classID, ownerAddress))
			require.Empty(t, target.ctx.EventManager().Events())

			defaultBytes, err := target.cdc.Marshal(nfttypes.DefaultGenesis())
			require.NoError(t, err)
			actual, err := target.keeper.ExportGenesis(target.ctx)
			require.NoError(t, err)
			actualBytes, err := target.cdc.Marshal(actual)
			require.NoError(t, err)
			require.Equal(t, defaultBytes, actualBytes)
		})
	}
}
