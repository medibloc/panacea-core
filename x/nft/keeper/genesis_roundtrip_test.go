package keeper

import (
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestBurnedStateGenesisRoundTrip(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner, _, data := createNFTForBurnTest(t, &fixture)
	secondRequest := validMintRequest(classID, controller, owner)
	secondRequest.NftId = "nft-2"
	secondRequest.Uri = "https://example.test/nft-2.json"
	secondRequest.UriHash = "sha256:" + strings.Repeat("c", 64)
	_, err := NewMsgServer(fixture.keeper).Mint(
		sdk.WrapSDKContext(fixture.ctx),
		secondRequest,
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		sdk.WrapSDKContext(fixture.ctx),
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)

	exported, err := fixture.keeper.ExportGenesis(fixture.ctx)
	require.NoError(t, err)
	require.Len(t, exported.Tombstones, 1)
	exportedBytes, err := fixture.cdc.Marshal(exported)
	require.NoError(t, err)
	var decodedGenesis nfttypes.GenesisState
	require.NoError(t, fixture.cdc.Unmarshal(exportedBytes, &decodedGenesis))
	require.Equal(t, data.TypeUrl, decodedGenesis.Tombstones[0].Data.TypeUrl)
	require.Equal(t, data.Value, decodedGenesis.Tombstones[0].Data.Value)
	require.IsType(
		t,
		&nfttypes.BasicNFTData{},
		decodedGenesis.Tombstones[0].Data.GetCachedValue(),
	)

	importedFixture := newKeeperFixture(t, true, true)
	require.NoError(t, importedFixture.keeper.InitGenesis(importedFixture.ctx, exported))
	require.Empty(t, importedFixture.ctx.EventManager().Events())
	importedTombstone, err := importedFixture.keeper.tombstones.Get(
		importedFixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	require.Equal(t, data.TypeUrl, importedTombstone.Data.TypeUrl)
	require.Equal(t, data.Value, importedTombstone.Data.Value)
	require.IsType(t, &nfttypes.BasicNFTData{}, importedTombstone.Data.GetCachedValue())
	importedLive, found := importedFixture.keeper.nftKeeper.GetNFT(
		importedFixture.ctx,
		classID,
		"nft-2",
	)
	require.True(t, found)
	require.Equal(t, secondRequest.Uri, importedLive.Uri)
	requireOwnerClassCount(t, &importedFixture, classID, owner, 1)
	assertClassNFTInvariants(t, &importedFixture, classID, 2, 1, 1)

	reexported, err := importedFixture.keeper.ExportGenesis(importedFixture.ctx)
	require.NoError(t, err)
	reexportedBytes, err := importedFixture.cdc.Marshal(reexported)
	require.NoError(t, err)
	require.Equal(t, exportedBytes, reexportedBytes)

	beforeRemint := snapshotRevokeState(t, &importedFixture, classID, "nft-1")
	remint := validMintRequest(classID, controller, owner)
	remint.Data = data
	_, err = NewMsgServer(importedFixture.keeper).Mint(
		sdk.WrapSDKContext(importedFixture.ctx),
		remint,
	)
	require.ErrorIs(t, err, nfttypes.ErrNFTIDPermanentlyUsed)
	require.Equal(
		t,
		beforeRemint,
		snapshotRevokeState(t, &importedFixture, classID, "nft-1"),
	)
	require.Empty(t, importedFixture.ctx.EventManager().Events())
}
