package keeper

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	upstreamnft "cosmossdk.io/x/nft"
	upstreamkeeper "cosmossdk.io/x/nft/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

func BenchmarkStandardQueryBalance(b *testing.B) {
	scenarios := []struct {
		ownerNFTs int
		classNFTs int
	}{
		{ownerNFTs: 1, classNFTs: 1},
		{ownerNFTs: 1, classNFTs: 1_000},
		{ownerNFTs: 10, classNFTs: 1_000},
		{ownerNFTs: 100, classNFTs: 1_000},
		{ownerNFTs: 1_000, classNFTs: 1_000},
	}

	for _, scenario := range scenarios {
		name := fmt.Sprintf("owner_%d/class_%d", scenario.ownerNFTs, scenario.classNFTs)
		b.Run(name, func(b *testing.B) {
			query := newBalanceQueryFixture(b, scenario.ownerNFTs, scenario.classNFTs, nil)
			response, err := query.server.Balance(query.goCtx, query.request)
			require.NoError(b, err)
			require.Equal(b, uint64(scenario.ownerNFTs), response.Amount)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response, err = query.server.Balance(query.goCtx, query.request)
				if err != nil {
					b.Fatal(err)
				}
				if response.Amount != uint64(scenario.ownerNFTs) {
					b.Fatalf("balance %d, expected %d", response.Amount, scenario.ownerNFTs)
				}
			}
		})
	}
}

func BenchmarkStandardQueryBalanceCountCandidates(b *testing.B) {
	scenarios := []struct {
		ownerNFTs int
		classNFTs int
	}{
		{ownerNFTs: 1, classNFTs: 1_000},
		{ownerNFTs: 1_000, classNFTs: 1_000},
	}

	for _, scenario := range scenarios {
		name := fmt.Sprintf("owner_%d/class_%d", scenario.ownerNFTs, scenario.classNFTs)
		b.Run(name, func(b *testing.B) {
			query := newBalanceQueryFixture(b, scenario.ownerNFTs, scenario.classNFTs, nil)

			b.Run("optimized_full_query", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					response, err := query.server.Balance(query.goCtx, query.request)
					if err != nil {
						b.Fatal(err)
					}
					if response.Amount != uint64(scenario.ownerNFTs) {
						b.Fatalf("balance %d, expected %d", response.Amount, scenario.ownerNFTs)
					}
				}
			})

			b.Run("owner_index_count_only", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					count, err := countBenchmarkOwnerIndex(
						query.ctx,
						query.keeper.nftStoreService,
						query.classID,
						query.owner,
					)
					if err != nil {
						b.Fatal(err)
					}
					if count != uint64(scenario.ownerNFTs) {
						b.Fatalf("balance %d, expected %d", count, scenario.ownerNFTs)
					}
				}
			})

			b.Run("derived_index_count_only", func(b *testing.B) {
				key := collections.Join(query.classID, query.request.Owner)
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					count, err := query.keeper.ownerClassCounts.Get(query.ctx, key)
					if err != nil {
						b.Fatal(err)
					}
					if count != uint64(scenario.ownerNFTs) {
						b.Fatalf("balance %d, expected %d", count, scenario.ownerNFTs)
					}
				}
			})
		})
	}
}

func TestStandardQueryBalanceStoreReadScaling(t *testing.T) {
	tests := []struct {
		name      string
		ownerNFTs int
		classNFTs int
	}{
		{
			name:      "one owned out of one",
			ownerNFTs: 1,
			classNFTs: 1,
		},
		{
			name:      "one owned out of one thousand",
			ownerNFTs: 1,
			classNFTs: 1_000,
		},
		{
			name:      "one thousand owned out of one thousand",
			ownerNFTs: 1_000,
			classNFTs: 1_000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counters := &queryStoreCounters{}
			query := newBalanceQueryFixture(t, test.ownerNFTs, test.classNFTs, counters)
			counters.reset()

			response, err := query.server.Balance(query.goCtx, query.request)
			require.NoError(t, err)
			require.Equal(t, uint64(test.ownerNFTs), response.Amount)
			require.Equal(t, uint64(1), counters.nft.gets)
			require.Zero(t, counters.nft.has)
			require.Zero(t, counters.nft.iterators)
			require.Zero(t, counters.nft.iteratorNexts)
			require.Equal(t, uint64(3), counters.policy.gets)
			require.Zero(t, counters.policy.has)
			require.Zero(t, counters.policy.iterators)
			require.Zero(t, counters.policy.iteratorNexts)
		})
	}
}

func TestStandardQueryBalanceCountCandidatesAgree(t *testing.T) {
	tests := []struct {
		ownerNFTs int
		classNFTs int
	}{
		{ownerNFTs: 1, classNFTs: 1_000},
		{ownerNFTs: 1_000, classNFTs: 1_000},
	}

	for _, test := range tests {
		name := fmt.Sprintf("owner_%d/class_%d", test.ownerNFTs, test.classNFTs)
		t.Run(name, func(t *testing.T) {
			query := newBalanceQueryFixture(t, test.ownerNFTs, test.classNFTs, nil)

			ownerIndexCount, err := countBenchmarkOwnerIndex(
				query.ctx,
				query.keeper.nftStoreService,
				query.classID,
				query.owner,
			)
			require.NoError(t, err)
			require.Equal(t, uint64(test.ownerNFTs), ownerIndexCount)

			derivedCount, err := query.keeper.ownerClassCounts.Get(
				query.ctx,
				collections.Join(query.classID, query.request.Owner),
			)
			require.NoError(t, err)
			require.Equal(t, uint64(test.ownerNFTs), derivedCount)

			response, err := query.server.Balance(query.goCtx, query.request)
			require.NoError(t, err)
			require.Equal(t, derivedCount, response.Amount)
		})
	}
}

type balanceQueryFixture struct {
	keeper  Keeper
	ctx     sdk.Context
	owner   sdk.AccAddress
	classID string
	server  upstreamnft.QueryServer
	goCtx   context.Context
	request *upstreamnft.QueryBalanceRequest
}

func newBalanceQueryFixture(
	t testing.TB,
	ownerNFTs int,
	classNFTs int,
	counters *queryStoreCounters,
) balanceQueryFixture {
	t.Helper()
	require.Positive(t, ownerNFTs)
	require.GreaterOrEqual(t, classNFTs, ownerNFTs)

	fixture := newKeeperFixture(t, true, true)
	if counters != nil {
		fixture.keeper = newQueryCountingKeeper(fixture, counters)
	}
	fixture.ctx = fixture.ctx.WithBlockTime(time.Date(2026, time.July, 23, 0, 0, 0, 0, time.UTC))

	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{91}, 20))
	ownerAddress := sdk.AccAddress(bytes.Repeat([]byte{92}, 20))
	otherOwnerAddress := sdk.AccAddress(bytes.Repeat([]byte{93}, 20))
	creator := fixture.accountAddress(t, creatorAddress)
	owner := fixture.accountAddress(t, ownerAddress)
	otherOwner := fixture.accountAddress(t, otherOwnerAddress)
	for _, address := range []sdk.AccAddress{creatorAddress, ownerAddress, otherOwnerAddress} {
		fixture.accountKeeper.accounts[string(address)] = authtypes.NewBaseAccountWithAddress(address)
	}

	createRequest := validCreateClassRequest(creator)
	createRequest.MaxSupply = 0
	created, err := NewMsgServer(fixture.keeper).CreateClass(
		sdk.WrapSDKContext(fixture.ctx),
		createRequest,
	)
	require.NoError(t, err)

	msgServer := NewMsgServer(fixture.keeper)
	for i := 0; i < classNFTs; i++ {
		recipient := otherOwner
		if i < ownerNFTs {
			recipient = owner
		}
		nftID := fmt.Sprintf("nft-%06d", i)
		mintRequest := validMintRequest(created.ClassId, creator, recipient)
		mintRequest.NftId = nftID
		mintRequest.Uri = "https://example.test/" + nftID + ".json"
		_, err := msgServer.Mint(sdk.WrapSDKContext(fixture.ctx), mintRequest)
		require.NoError(t, err)
	}
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())

	return balanceQueryFixture{
		keeper:  fixture.keeper,
		ctx:     fixture.ctx,
		owner:   ownerAddress,
		classID: created.ClassId,
		server:  NewStandardQueryServer(fixture.keeper),
		goCtx:   sdk.WrapSDKContext(fixture.ctx),
		request: &upstreamnft.QueryBalanceRequest{
			ClassId: created.ClassId,
			Owner:   owner,
		},
	}
}

func countBenchmarkOwnerIndex(
	ctx context.Context,
	storeService corestore.KVStoreService,
	classID string,
	owner sdk.AccAddress,
) (uint64, error) {
	// This intentionally mirrors the upstream private
	// nftOfClassByOwnerStoreKey encoding to measure that candidate's cost.
	lengthPrefixedOwner := address.MustLengthPrefix(owner)
	prefix := make(
		[]byte,
		0,
		len(upstreamkeeper.NFTOfClassByOwnerKey)+
			len(lengthPrefixedOwner)+
			len(upstreamkeeper.Delimiter)+
			len(classID)+
			len(upstreamkeeper.Delimiter),
	)
	prefix = append(prefix, upstreamkeeper.NFTOfClassByOwnerKey...)
	prefix = append(prefix, lengthPrefixedOwner...)
	prefix = append(prefix, upstreamkeeper.Delimiter...)
	prefix = append(prefix, classID...)
	prefix = append(prefix, upstreamkeeper.Delimiter...)

	store := storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return 0, err
	}
	var count uint64
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	if err := iterator.Close(); err != nil {
		return 0, err
	}
	return count, nil
}
