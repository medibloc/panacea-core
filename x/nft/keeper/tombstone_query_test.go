package keeper

import (
	"strings"
	"testing"
	"time"

	"cosmossdk.io/collections"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNFTRecordReturnsBurnTombstone(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
	fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
	_, err := NewMsgServer(fixture.keeper).Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)
	tombstone, err := fixture.keeper.tombstones.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)

	queryResponse, err := NewQueryServer(fixture.keeper).NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "nft-1"},
	)
	require.NoError(t, err)
	require.Nil(t, queryResponse.NftRecord.GetLive())
	queriedTombstone := queryResponse.NftRecord.GetBurnTombstone()
	require.NotNil(t, queriedTombstone)
	require.Equal(t, tombstone.ClassId, queriedTombstone.ClassId)
	require.Equal(t, tombstone.NftId, queriedTombstone.NftId)
	require.Equal(t, tombstone.Mint, queriedTombstone.Mint)
	require.Equal(t, tombstone.Uri, queriedTombstone.Uri)
	require.Equal(t, tombstone.UriHash, queriedTombstone.UriHash)
	require.Equal(t, tombstone.Data.TypeUrl, queriedTombstone.Data.TypeUrl)
	require.Equal(t, tombstone.Data.Value, queriedTombstone.Data.Value)
	require.Nil(t, queriedTombstone.Data.GetCachedValue())
	require.Equal(t, tombstone.BurnedAt, queriedTombstone.BurnedAt)
	require.Equal(t, tombstone.BurnedBy, queriedTombstone.BurnedBy)

	responseBytes, err := fixture.cdc.Marshal(queryResponse)
	require.NoError(t, err)
	var decodedResponse nfttypes.QueryNFTRecordResponse
	require.NoError(t, fixture.cdc.Unmarshal(responseBytes, &decodedResponse))
	require.IsType(
		t,
		&nfttypes.BasicNFTData{},
		decodedResponse.NftRecord.GetBurnTombstone().Data.GetCachedValue(),
	)
}

func TestNFTRecordReturnsBurnedRevocation(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, controller, owner, _, _ := createNFTForBurnTest(t, &fixture)
	revokedAt := fixture.ctx.BlockTime().Add(time.Hour)
	fixture.ctx = fixture.ctx.WithBlockTime(revokedAt)
	_, err := NewMsgServer(fixture.keeper).Revoke(
		fixture.ctx,
		&nfttypes.MsgRevokeRequest{
			ClassId: classID, NftId: "nft-1", Controller: controller,
		},
	)
	require.NoError(t, err)
	lifecycle, err := fixture.keeper.lifecycles.Get(
		fixture.ctx,
		collections.Join(classID, "nft-1"),
	)
	require.NoError(t, err)
	fixture.ctx = fixture.ctx.WithBlockTime(revokedAt.Add(time.Hour))
	_, err = NewMsgServer(fixture.keeper).Burn(
		fixture.ctx,
		&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
	)
	require.NoError(t, err)

	queryResponse, err := NewQueryServer(fixture.keeper).NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "nft-1"},
	)
	require.NoError(t, err)
	require.Equal(t, lifecycle.Revocation, queryResponse.NftRecord.GetBurnTombstone().Revocation)
}

func TestNFTRecordRejectsInvalidTombstone(t *testing.T) {
	for _, tc := range []struct {
		name          string
		expectedError string
		mutate        func(tombstone *nfttypes.BurnTombstone)
	}{
		{
			name:          "missing mint record",
			expectedError: "no mint record",
			mutate:        func(tombstone *nfttypes.BurnTombstone) { tombstone.Mint = nil },
		},
		{
			name:          "burn predates mint",
			expectedError: "burn predates mint",
			mutate: func(tombstone *nfttypes.BurnTombstone) {
				tombstone.BurnedAt = tombstone.Mint.MintedAt.Add(-time.Second)
			},
		},
		{
			name:          "zero burn time",
			expectedError: "no burn time",
			mutate:        func(tombstone *nfttypes.BurnTombstone) { tombstone.BurnedAt = time.Time{} },
		},
		{
			name:          "key and value mismatch",
			expectedError: "key does not match value",
			mutate:        func(tombstone *nfttypes.BurnTombstone) { tombstone.NftId = "other" },
		},
		{
			name:          "invalid minter",
			expectedError: "invalid minter",
			mutate:        func(tombstone *nfttypes.BurnTombstone) { tombstone.Mint.MintedBy = "invalid" },
		},
		{
			name:          "non-canonical minter",
			expectedError: "minter is not canonical",
			mutate: func(tombstone *nfttypes.BurnTombstone) {
				tombstone.Mint.MintedBy = strings.ToUpper(tombstone.Mint.MintedBy)
			},
		},
		{
			name:          "revocation predates mint",
			expectedError: "revocation predates mint",
			mutate: func(tombstone *nfttypes.BurnTombstone) {
				tombstone.Revocation = &nfttypes.Revocation{
					RevokedAt: tombstone.Mint.MintedAt.Add(-time.Second),
					RevokedBy: tombstone.Mint.MintedBy,
				}
			},
		},
		{
			name:          "burn predates revocation",
			expectedError: "burn predates revocation",
			mutate: func(tombstone *nfttypes.BurnTombstone) {
				tombstone.Revocation = &nfttypes.Revocation{
					RevokedAt: tombstone.BurnedAt.Add(time.Second),
					RevokedBy: tombstone.Mint.MintedBy,
				}
			},
		},
		{
			name:          "invalid burner",
			expectedError: "invalid burner",
			mutate:        func(tombstone *nfttypes.BurnTombstone) { tombstone.BurnedBy = "invalid" },
		},
		{
			name:          "non-canonical burner",
			expectedError: "burner is not canonical",
			mutate: func(tombstone *nfttypes.BurnTombstone) {
				tombstone.BurnedBy = strings.ToUpper(tombstone.BurnedBy)
			},
		},
		{
			name:          "invalid URI metadata",
			expectedError: "invalid stored URI metadata",
			mutate:        func(tombstone *nfttypes.BurnTombstone) { tombstone.Uri = "" },
		},
		{
			name:          "invalid metadata",
			expectedError: "data must contain at least one field",
			mutate: func(tombstone *nfttypes.BurnTombstone) {
				empty, err := cdctypes.NewAnyWithValue(&nfttypes.BasicNFTData{})
				require.NoError(t, err)
				tombstone.Data = empty
			},
		},
		{
			name:          "unknown metadata type",
			expectedError: "no concrete type registered for type URL",
			mutate: func(tombstone *nfttypes.BurnTombstone) {
				tombstone.Data = &cdctypes.Any{
					TypeUrl: "/panacea.nft.v1.UnknownNFTData",
					Value:   []byte{0x0a, 0x00},
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, _, owner, _, _ := createNFTForBurnTest(t, &fixture)
			fixture.ctx = fixture.ctx.WithBlockTime(fixture.ctx.BlockTime().Add(time.Hour))
			_, err := NewMsgServer(fixture.keeper).Burn(
				fixture.ctx,
				&nfttypes.MsgBurnRequest{ClassId: classID, NftId: "nft-1", Owner: owner},
			)
			require.NoError(t, err)
			key := collections.Join(classID, "nft-1")
			tombstone, err := fixture.keeper.tombstones.Get(fixture.ctx, key)
			require.NoError(t, err)
			tc.mutate(&tombstone)
			require.NoError(t, fixture.keeper.tombstones.Set(fixture.ctx, key, tombstone))

			_, err = NewQueryServer(fixture.keeper).NFTRecord(
				fixture.ctx,
				&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "nft-1"},
			)
			require.Equal(t, codes.Internal, status.Code(err))
			require.ErrorContains(t, err, tc.expectedError)
		})
	}
}

func TestNFTRecordReturnsNotFoundForUnknownID(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classID, _, _, _, _ := createNFTForBurnTest(t, &fixture)

	_, err := NewQueryServer(fixture.keeper).NFTRecord(
		fixture.ctx,
		&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "unknown"},
	)
	require.Equal(t, codes.NotFound, status.Code(err))
}
