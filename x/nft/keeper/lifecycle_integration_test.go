package keeper

import (
	"bytes"
	"testing"
	"time"

	upstreamnft "cosmossdk.io/x/nft"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFullNFTLifecycleAcrossPanaceaAndStandardServices(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	panaceaMsg := NewMsgServer(fixture.keeper)
	standardMsg := NewStandardMsgServer(fixture.keeper)
	panaceaQuery := NewQueryServer(fixture.keeper)
	standardQuery := NewStandardQueryServer(fixture.keeper)

	creator := registerLifecycleAccount(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{121}, 20)),
	)
	controller := registerLifecycleAccount(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{122}, 20)),
	)
	firstOwner := registerLifecycleAccount(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{123}, 20)),
	)
	finalOwner := registerLifecycleAccount(
		t,
		&fixture,
		sdk.AccAddress(bytes.Repeat([]byte{124}, 32)),
	)

	baseTime := time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC)
	fixture.ctx = fixture.ctx.WithBlockTime(baseTime)

	// Create the class with the creator as its initial controller.
	classRequest := validCreateClassRequest(creator)
	classResponse, err := panaceaMsg.CreateClass(
		fixture.ctx,
		classRequest,
	)
	require.NoError(t, err)
	classID := classResponse.ClassId

	classRecordResponse, err := panaceaQuery.ClassRecord(
		fixture.ctx,
		&nfttypes.QueryClassRecordRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.NotNil(t, classRecordResponse.ClassRecord)
	require.NotNil(t, classRecordResponse.ClassRecord.Policy)
	require.Equal(t, creator, classRecordResponse.ClassRecord.Policy.Creator)
	require.Equal(t, creator, classRecordResponse.ClassRecord.Policy.Controller)
	require.Equal(t, classRequest.TransferPolicy, classRecordResponse.ClassRecord.Policy.TransferPolicy)
	require.True(t, classRecordResponse.ClassRecord.Policy.Revocable)
	require.Equal(t, classRequest.MaxSupply, classRecordResponse.ClassRecord.Policy.MaxSupply)
	require.Zero(t, classRecordResponse.ClassRecord.MintedCount)

	standardClass, err := standardQuery.Class(
		fixture.ctx,
		&upstreamnft.QueryClassRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.NotNil(t, standardClass.Class)
	require.Equal(t, classID, standardClass.Class.Id)
	standardSupply, err := standardQuery.Supply(
		fixture.ctx,
		&upstreamnft.QuerySupplyRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Zero(t, standardSupply.Amount)
	assertClassNFTInvariants(t, &fixture, classID, 0, 0, 0)

	// Transfer class control before minting.
	fixture.ctx = fixture.ctx.
		WithBlockTime(baseTime.Add(time.Hour)).
		WithEventManager(sdk.NewEventManager())
	_, err = panaceaMsg.UpdateController(
		fixture.ctx,
		&nfttypes.MsgUpdateControllerRequest{
			ClassId:       classID,
			Controller:    creator,
			NewController: controller,
		},
	)
	require.NoError(t, err)
	classRecordResponse, err = panaceaQuery.ClassRecord(
		fixture.ctx,
		&nfttypes.QueryClassRecordRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Equal(t, creator, classRecordResponse.ClassRecord.Policy.Creator)
	require.Equal(t, controller, classRecordResponse.ClassRecord.Policy.Controller)
	require.Zero(t, classRecordResponse.ClassRecord.MintedCount)

	// Mint one ACTIVE NFT through the Panacea service.
	metadata, err := codectypes.NewAnyWithValue(&nfttypes.BasicNFTData{
		Name:        "Certificate #1",
		Description: "Lifecycle integration certificate",
		ImageUri:    "https://example.test/certificate-1.png",
	})
	require.NoError(t, err)
	mintedAt := baseTime.Add(2 * time.Hour)
	fixture.ctx = fixture.ctx.
		WithBlockTime(mintedAt).
		WithEventManager(sdk.NewEventManager())
	mintRequest := validMintRequest(classID, controller, firstOwner)
	mintRequest.Data = metadata
	_, err = panaceaMsg.Mint(fixture.ctx, mintRequest)
	require.NoError(t, err)

	nftRecordResponse, err := panaceaQuery.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: mintRequest.NftId},
	)
	require.NoError(t, err)
	live := nftRecordResponse.NftRecord.GetLive()
	require.NotNil(t, live)
	require.Equal(t, firstOwner, live.Owner)
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_ACTIVE, live.Status)
	require.Equal(t, &nfttypes.MintRecord{
		MintedAt: mintedAt,
		MintedBy: controller,
	}, live.Mint)
	require.Nil(t, live.Revocation)
	require.NotNil(t, live.Nft)
	require.NotNil(t, live.Nft.Data)
	require.Equal(t, metadata.TypeUrl, live.Nft.Data.TypeUrl)
	require.Equal(t, metadata.Value, live.Nft.Data.Value)
	mintRecord := live.Mint

	standardNFT, err := standardQuery.NFT(
		fixture.ctx,
		&upstreamnft.QueryNFTRequest{ClassId: classID, Id: mintRequest.NftId},
	)
	require.NoError(t, err)
	require.Equal(t, mintRequest.Uri, standardNFT.Nft.Uri)
	standardOwner, err := standardQuery.Owner(
		fixture.ctx,
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: mintRequest.NftId},
	)
	require.NoError(t, err)
	require.Equal(t, firstOwner, standardOwner.Owner)
	standardSupply, err = standardQuery.Supply(
		fixture.ctx,
		&upstreamnft.QuerySupplyRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Equal(t, uint64(1), standardSupply.Amount)
	assertClassNFTInvariants(t, &fixture, classID, 1, 1, 0)

	// Transfer ownership through the policy-aware standard MsgSend.
	fixture.ctx = fixture.ctx.
		WithBlockTime(baseTime.Add(3 * time.Hour)).
		WithEventManager(sdk.NewEventManager())
	_, err = standardMsg.Send(
		fixture.ctx,
		&upstreamnft.MsgSend{
			ClassId:  classID,
			Id:       mintRequest.NftId,
			Sender:   firstOwner,
			Receiver: finalOwner,
		},
	)
	require.NoError(t, err)
	nftRecordResponse, err = panaceaQuery.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: mintRequest.NftId},
	)
	require.NoError(t, err)
	live = nftRecordResponse.NftRecord.GetLive()
	require.NotNil(t, live)
	require.Equal(t, finalOwner, live.Owner)
	require.Equal(t, mintRecord, live.Mint)
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_ACTIVE, live.Status)
	standardOwner, err = standardQuery.Owner(
		fixture.ctx,
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: mintRequest.NftId},
	)
	require.NoError(t, err)
	require.Equal(t, finalOwner, standardOwner.Owner)
	require.Equal(t, uint64(0), queryStandardBalance(t, &fixture, standardQuery, classID, firstOwner))
	require.Equal(t, uint64(1), queryStandardBalance(t, &fixture, standardQuery, classID, finalOwner))
	assertClassNFTInvariants(t, &fixture, classID, 1, 1, 0)

	// Revoke without changing owner, supply, mint provenance, or standard NFT data.
	revokedAt := baseTime.Add(4 * time.Hour)
	fixture.ctx = fixture.ctx.
		WithBlockTime(revokedAt).
		WithEventManager(sdk.NewEventManager())
	_, err = panaceaMsg.Revoke(
		fixture.ctx,
		&nfttypes.MsgRevokeRequest{
			ClassId:    classID,
			NftId:      mintRequest.NftId,
			Controller: controller,
		},
	)
	require.NoError(t, err)
	nftRecordResponse, err = panaceaQuery.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: mintRequest.NftId},
	)
	require.NoError(t, err)
	live = nftRecordResponse.NftRecord.GetLive()
	require.NotNil(t, live)
	require.Equal(t, finalOwner, live.Owner)
	require.Equal(t, mintRecord, live.Mint)
	require.Equal(t, nfttypes.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED, live.Status)
	require.Equal(t, &nfttypes.Revocation{
		RevokedAt: revokedAt,
		RevokedBy: controller,
	}, live.Revocation)
	revocation := live.Revocation

	standardNFT, err = standardQuery.NFT(
		fixture.ctx,
		&upstreamnft.QueryNFTRequest{ClassId: classID, Id: mintRequest.NftId},
	)
	require.NoError(t, err)
	require.NotNil(t, standardNFT.Nft)
	standardOwner, err = standardQuery.Owner(
		fixture.ctx,
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: mintRequest.NftId},
	)
	require.NoError(t, err)
	require.Equal(t, finalOwner, standardOwner.Owner)
	assertClassNFTInvariants(t, &fixture, classID, 1, 1, 0)

	// Burn as the current owner and preserve both mint and revocation provenance.
	burnedAt := baseTime.Add(5 * time.Hour)
	fixture.ctx = fixture.ctx.
		WithBlockTime(burnedAt).
		WithEventManager(sdk.NewEventManager())
	_, err = panaceaMsg.Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{
			ClassId: classID,
			NftId:   mintRequest.NftId,
			Owner:   finalOwner,
		},
	)
	require.NoError(t, err)

	nftRecordResponse, err = panaceaQuery.NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: mintRequest.NftId},
	)
	require.NoError(t, err)
	require.Nil(t, nftRecordResponse.NftRecord.GetLive())
	tombstone := nftRecordResponse.NftRecord.GetBurnTombstone()
	require.NotNil(t, tombstone)
	require.NotNil(t, tombstone.Data)
	require.Equal(t, mintRecord, tombstone.Mint)
	require.Equal(t, revocation, tombstone.Revocation)
	require.Equal(t, mintRequest.Uri, tombstone.Uri)
	require.Equal(t, mintRequest.UriHash, tombstone.UriHash)
	require.Equal(t, metadata.TypeUrl, tombstone.Data.TypeUrl)
	require.Equal(t, metadata.Value, tombstone.Data.Value)
	require.Equal(t, burnedAt, tombstone.BurnedAt)
	require.Equal(t, finalOwner, tombstone.BurnedBy)

	_, err = standardQuery.NFT(
		fixture.ctx,
		&upstreamnft.QueryNFTRequest{ClassId: classID, Id: mintRequest.NftId},
	)
	require.Equal(t, codes.NotFound, status.Code(err))
	standardOwner, err = standardQuery.Owner(
		fixture.ctx,
		&upstreamnft.QueryOwnerRequest{ClassId: classID, Id: mintRequest.NftId},
	)
	require.NoError(t, err)
	require.Empty(t, standardOwner.Owner)
	standardSupply, err = standardQuery.Supply(
		fixture.ctx,
		&upstreamnft.QuerySupplyRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Zero(t, standardSupply.Amount)
	require.Zero(t, queryStandardBalance(t, &fixture, standardQuery, classID, firstOwner))
	require.Zero(t, queryStandardBalance(t, &fixture, standardQuery, classID, finalOwner))

	classRecordResponse, err = panaceaQuery.ClassRecord(
		fixture.ctx,
		&nfttypes.QueryClassRecordRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Equal(t, controller, classRecordResponse.ClassRecord.Policy.Controller)
	require.Equal(t, uint64(1), classRecordResponse.ClassRecord.MintedCount)
	liveRecords, err := panaceaQuery.NFTRecords(
		fixture.ctx,
		&nfttypes.QueryNFTRecordsRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.Empty(t, liveRecords.NftRecords)
	require.NotNil(t, liveRecords.Pagination)
	assertClassNFTInvariants(t, &fixture, classID, 1, 0, 1)
}

func registerLifecycleAccount(
	t *testing.T,
	fixture *keeperFixture,
	address sdk.AccAddress,
) string {
	t.Helper()
	fixture.accountKeeper.accounts[string(address)] = authtypes.NewBaseAccountWithAddress(address)
	return fixture.accountAddress(t, address)
}

func queryStandardBalance(
	t *testing.T,
	fixture *keeperFixture,
	queryServer upstreamnft.QueryServer,
	classID string,
	owner string,
) uint64 {
	t.Helper()
	response, err := queryServer.Balance(
		fixture.ctx,
		&upstreamnft.QueryBalanceRequest{ClassId: classID, Owner: owner},
	)
	require.NoError(t, err)
	return response.Amount
}
