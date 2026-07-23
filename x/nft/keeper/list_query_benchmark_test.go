package keeper

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

// One item beyond the page limit ensures the measured page also produces a
// continuation key.
const maximumPageBenchmarkStateSize = int(maximumQueryPageLimit) + 1

func BenchmarkMaximumPageListQueries(b *testing.B) {
	b.Run("standard_classes", func(b *testing.B) {
		fixture := newMaximumPageClassesFixture(b, nil)
		request := &upstreamnft.QueryClassesRequest{
			Pagination: &query.PageRequest{Limit: maximumQueryPageLimit},
		}
		response, err := fixture.standard.Classes(fixture.goCtx, request)
		require.NoError(b, err)
		require.Len(b, response.Classes, int(maximumQueryPageLimit))
		require.NotEmpty(b, response.Pagination.NextKey)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			response, err = fixture.standard.Classes(fixture.goCtx, request)
			if err != nil {
				b.Fatal(err)
			}
			if len(response.Classes) != int(maximumQueryPageLimit) {
				b.Fatalf(
					"classes length %d, expected %d",
					len(response.Classes),
					maximumQueryPageLimit,
				)
			}
		}
	})

	for _, filter := range maximumPageNFTFilters() {
		b.Run("standard_nfts/"+filter.name, func(b *testing.B) {
			fixture := newMaximumPageNFTFixture(b, nil)
			request := filter.standardRequest(fixture)
			response, err := fixture.standard.NFTs(fixture.goCtx, request)
			require.NoError(b, err)
			require.Len(b, response.Nfts, int(maximumQueryPageLimit))
			require.NotEmpty(b, response.Pagination.NextKey)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response, err = fixture.standard.NFTs(fixture.goCtx, request)
				if err != nil {
					b.Fatal(err)
				}
				if len(response.Nfts) != int(maximumQueryPageLimit) {
					b.Fatalf(
						"nfts length %d, expected %d",
						len(response.Nfts),
						maximumQueryPageLimit,
					)
				}
			}
		})

		b.Run("panacea_nft_records/"+filter.name, func(b *testing.B) {
			fixture := newMaximumPageNFTFixture(b, nil)
			request := filter.panaceaRequest(fixture)
			response, err := fixture.panacea.NFTRecords(fixture.goCtx, request)
			require.NoError(b, err)
			require.Len(b, response.NftRecords, int(maximumQueryPageLimit))
			require.NotEmpty(b, response.Pagination.NextKey)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response, err = fixture.panacea.NFTRecords(fixture.goCtx, request)
				if err != nil {
					b.Fatal(err)
				}
				if len(response.NftRecords) != int(maximumQueryPageLimit) {
					b.Fatalf(
						"nft records length %d, expected %d",
						len(response.NftRecords),
						maximumQueryPageLimit,
					)
				}
			}
		})
	}
}

func TestMaximumPageListQueryStoreReads(t *testing.T) {
	t.Run("standard classes", func(t *testing.T) {
		counters := &queryStoreCounters{}
		fixture := newMaximumPageClassesFixture(t, counters)
		counters.reset()

		response, err := fixture.standard.Classes(
			fixture.goCtx,
			&upstreamnft.QueryClassesRequest{
				Pagination: &query.PageRequest{Limit: maximumQueryPageLimit},
			},
		)
		require.NoError(t, err)
		require.Len(t, response.Classes, int(maximumQueryPageLimit))
		require.NotEmpty(t, response.Pagination.NextKey)
		require.Equal(t, storeReadCounters{
			gets:          100,
			iterators:     1,
			iteratorNexts: 100,
		}, counters.nft)
		require.Equal(t, storeReadCounters{gets: 200}, counters.policy)
	})

	fixtureCounters := &queryStoreCounters{}
	fixture := newMaximumPageNFTFixture(t, fixtureCounters)
	for _, filter := range maximumPageNFTFilters() {
		t.Run("standard nfts/"+filter.name, func(t *testing.T) {
			fixtureCounters.reset()
			response, err := fixture.standard.NFTs(
				fixture.goCtx,
				filter.standardRequest(fixture),
			)
			require.NoError(t, err)
			require.Len(t, response.Nfts, int(maximumQueryPageLimit))
			require.NotEmpty(t, response.Pagination.NextKey)
			require.Equal(t, filter.expectedNFTReads, fixtureCounters.nft)
			require.Equal(t, filter.expectedPolicyReads, fixtureCounters.policy)
		})

		t.Run("panacea nft records/"+filter.name, func(t *testing.T) {
			fixtureCounters.reset()
			response, err := fixture.panacea.NFTRecords(
				fixture.goCtx,
				filter.panaceaRequest(fixture),
			)
			require.NoError(t, err)
			require.Len(t, response.NftRecords, int(maximumQueryPageLimit))
			require.NotEmpty(t, response.Pagination.NextKey)
			require.Equal(t, filter.expectedNFTReads, fixtureCounters.nft)
			require.Equal(t, filter.expectedPolicyReads, fixtureCounters.policy)
		})
	}
}

type maximumPageQueryFixture struct {
	goCtx    context.Context
	classID  string
	owner    string
	standard upstreamnft.QueryServer
	panacea  nfttypes.QueryServer
}

func newMaximumPageClassesFixture(
	t testing.TB,
	counters *queryStoreCounters,
) maximumPageQueryFixture {
	t.Helper()
	fixture := newKeeperFixture(t, true, true)
	if counters != nil {
		fixture.keeper = newQueryCountingKeeper(fixture, counters)
	}
	fixture.ctx = fixture.ctx.WithBlockTime(
		time.Date(2026, time.July, 23, 0, 0, 0, 0, time.UTC),
	)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{112}, 20))
	creator := fixture.accountAddress(t, creatorAddress)
	fixture.accountKeeper.accounts[string(creatorAddress)] =
		authtypes.NewBaseAccountWithAddress(creatorAddress)
	server := NewMsgServer(fixture.keeper)
	for index := 0; index < maximumPageBenchmarkStateSize; index++ {
		request := validCreateClassRequest(creator)
		request.LocalClassId = fmt.Sprintf("benchmark%03d", index)
		_, err := server.CreateClass(sdk.WrapSDKContext(fixture.ctx), request)
		require.NoError(t, err)
	}
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return maximumPageQueryFixture{
		goCtx:    sdk.WrapSDKContext(fixture.ctx),
		standard: NewStandardQueryServer(fixture.keeper),
		panacea:  NewQueryServer(fixture.keeper),
	}
}

func newMaximumPageNFTFixture(
	t testing.TB,
	counters *queryStoreCounters,
) maximumPageQueryFixture {
	t.Helper()
	fixture := newKeeperFixture(t, true, true)
	if counters != nil {
		fixture.keeper = newQueryCountingKeeper(fixture, counters)
	}
	fixture.ctx = fixture.ctx.WithBlockTime(
		time.Date(2026, time.July, 23, 0, 0, 0, 0, time.UTC),
	)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{113}, 20))
	creator := fixture.accountAddress(t, creatorAddress)
	ownerAddress := sdk.AccAddress(bytes.Repeat([]byte{114}, 32))
	owner := fixture.accountAddress(t, ownerAddress)
	for _, address := range []sdk.AccAddress{creatorAddress, ownerAddress} {
		fixture.accountKeeper.accounts[string(address)] =
			authtypes.NewBaseAccountWithAddress(address)
	}

	classRequest := validCreateClassRequest(creator)
	classRequest.LocalClassId = "benchmarknfts"
	classRequest.MaxSupply = 0
	classResponse, err := NewMsgServer(fixture.keeper).CreateClass(
		sdk.WrapSDKContext(fixture.ctx),
		classRequest,
	)
	require.NoError(t, err)
	server := NewMsgServer(fixture.keeper)
	for index := 0; index < maximumPageBenchmarkStateSize; index++ {
		nftID := fmt.Sprintf("nft-%03d", index)
		request := validMintRequest(classResponse.ClassId, creator, owner)
		request.NftId = nftID
		request.Uri = "https://example.test/" + nftID + ".json"
		_, err := server.Mint(sdk.WrapSDKContext(fixture.ctx), request)
		require.NoError(t, err)
	}
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return maximumPageQueryFixture{
		goCtx:    sdk.WrapSDKContext(fixture.ctx),
		classID:  classResponse.ClassId,
		owner:    owner,
		standard: NewStandardQueryServer(fixture.keeper),
		panacea:  NewQueryServer(fixture.keeper),
	}
}

type maximumPageNFTFilter struct {
	name                string
	expectedNFTReads    storeReadCounters
	expectedPolicyReads storeReadCounters
	standardRequest     func(maximumPageQueryFixture) *upstreamnft.QueryNFTsRequest
	panaceaRequest      func(maximumPageQueryFixture) *nfttypes.QueryNFTRecordsRequest
}

func maximumPageNFTFilters() []maximumPageNFTFilter {
	page := func() *query.PageRequest {
		return &query.PageRequest{Limit: maximumQueryPageLimit}
	}
	return []maximumPageNFTFilter{
		{
			name: "class",
			expectedNFTReads: storeReadCounters{
				gets:          202,
				iterators:     1,
				iteratorNexts: 100,
			},
			expectedPolicyReads: storeReadCounters{gets: 202},
			standardRequest: func(fixture maximumPageQueryFixture) *upstreamnft.QueryNFTsRequest {
				return &upstreamnft.QueryNFTsRequest{
					ClassId:    fixture.classID,
					Pagination: page(),
				}
			},
			panaceaRequest: func(fixture maximumPageQueryFixture) *nfttypes.QueryNFTRecordsRequest {
				return &nfttypes.QueryNFTRecordsRequest{
					ClassId:    fixture.classID,
					Pagination: page(),
				}
			},
		},
		{
			name: "owner",
			expectedNFTReads: storeReadCounters{
				gets:          302,
				iterators:     1,
				iteratorNexts: 100,
			},
			expectedPolicyReads: storeReadCounters{gets: 202},
			standardRequest: func(fixture maximumPageQueryFixture) *upstreamnft.QueryNFTsRequest {
				return &upstreamnft.QueryNFTsRequest{
					Owner:      fixture.owner,
					Pagination: page(),
				}
			},
			panaceaRequest: func(fixture maximumPageQueryFixture) *nfttypes.QueryNFTRecordsRequest {
				return &nfttypes.QueryNFTRecordsRequest{
					Owner:      fixture.owner,
					Pagination: page(),
				}
			},
		},
		{
			name: "class_owner",
			expectedNFTReads: storeReadCounters{
				gets:          302,
				iterators:     1,
				iteratorNexts: 100,
			},
			expectedPolicyReads: storeReadCounters{gets: 202},
			standardRequest: func(fixture maximumPageQueryFixture) *upstreamnft.QueryNFTsRequest {
				return &upstreamnft.QueryNFTsRequest{
					ClassId:    fixture.classID,
					Owner:      fixture.owner,
					Pagination: page(),
				}
			},
			panaceaRequest: func(fixture maximumPageQueryFixture) *nfttypes.QueryNFTRecordsRequest {
				return &nfttypes.QueryNFTRecordsRequest{
					ClassId:    fixture.classID,
					Owner:      fixture.owner,
					Pagination: page(),
				}
			},
		},
	}
}
