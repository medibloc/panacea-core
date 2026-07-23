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
	"github.com/cosmos/cosmos-sdk/runtime"
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
			counts := newBenchmarkBalanceCounts(b, query)

			b.Run("current_full_query", func(b *testing.B) {
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
					count, err := counts.Get(query.ctx, key)
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
		name          string
		ownerNFTs     int
		classNFTs     int
		expectedGets  uint64
		expectedIters uint64
		expectedNexts uint64
	}{
		{
			name:          "one owned out of one",
			ownerNFTs:     1,
			classNFTs:     1,
			expectedGets:  2,
			expectedIters: 1,
			expectedNexts: 1,
		},
		{
			name:          "one owned out of one thousand",
			ownerNFTs:     1,
			classNFTs:     1_000,
			expectedGets:  2,
			expectedIters: 1,
			expectedNexts: 1,
		},
		{
			name:          "one thousand owned out of one thousand",
			ownerNFTs:     1_000,
			classNFTs:     1_000,
			expectedGets:  1_001,
			expectedIters: 1,
			expectedNexts: 1_000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counters := &balanceStoreCounters{}
			query := newBalanceQueryFixture(t, test.ownerNFTs, test.classNFTs, counters)
			counters.reset()

			response, err := query.server.Balance(query.goCtx, query.request)
			require.NoError(t, err)
			require.Equal(t, uint64(test.ownerNFTs), response.Amount)
			require.Equal(t, test.expectedGets, counters.gets)
			require.Equal(t, test.expectedIters, counters.iterators)
			require.Equal(t, test.expectedNexts, counters.iteratorNexts)
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
			counts := newBenchmarkBalanceCounts(t, query)

			ownerIndexCount, err := countBenchmarkOwnerIndex(
				query.ctx,
				query.keeper.nftStoreService,
				query.classID,
				query.owner,
			)
			require.NoError(t, err)
			require.Equal(t, uint64(test.ownerNFTs), ownerIndexCount)

			derivedCount, err := counts.Get(
				query.ctx,
				collections.Join(query.classID, query.request.Owner),
			)
			require.NoError(t, err)
			require.Equal(t, uint64(test.ownerNFTs), derivedCount)
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
	counters *balanceStoreCounters,
) balanceQueryFixture {
	t.Helper()
	require.Positive(t, ownerNFTs)
	require.GreaterOrEqual(t, classNFTs, ownerNFTs)

	fixture := newKeeperFixture(t, true, true)
	if counters != nil {
		nftService := balanceCountingStoreService{
			delegate: runtime.NewKVStoreService(fixture.nftService),
			counters: counters,
		}
		fixture.keeper = NewKeeper(
			fixture.cdc,
			nftService,
			runtime.NewKVStoreService(fixture.policyService),
			fixture.accountKeeper,
			testBankKeeper{},
			fixture.moduleAccountAddresses,
		)
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

func newBenchmarkBalanceCounts(
	t testing.TB,
	query balanceQueryFixture,
) collections.Map[collections.Pair[string, string], uint64] {
	t.Helper()
	builder := collections.NewSchemaBuilder(query.keeper.policyStoreService)
	counts := collections.NewMap(
		builder,
		collections.NewPrefix(255),
		"benchmark_balance_counts",
		nftKeyCodec,
		collections.Uint64Value,
	)
	_, err := builder.Build()
	require.NoError(t, err)
	require.NoError(t, counts.Set(
		query.ctx,
		collections.Join(query.classID, query.request.Owner),
		uint64(query.keeper.nftKeeper.GetBalance(query.ctx, query.classID, query.owner)),
	))
	return counts
}

type balanceStoreCounters struct {
	gets          uint64
	iterators     uint64
	iteratorNexts uint64
}

func (c *balanceStoreCounters) reset() {
	c.gets = 0
	c.iterators = 0
	c.iteratorNexts = 0
}

type balanceCountingStoreService struct {
	delegate corestore.KVStoreService
	counters *balanceStoreCounters
}

func (s balanceCountingStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return balanceCountingStore{
		KVStore:  s.delegate.OpenKVStore(ctx),
		counters: s.counters,
	}
}

type balanceCountingStore struct {
	corestore.KVStore
	counters *balanceStoreCounters
}

func (s balanceCountingStore) Get(key []byte) ([]byte, error) {
	s.counters.gets++
	return s.KVStore.Get(key)
}

func (s balanceCountingStore) Iterator(start, end []byte) (corestore.Iterator, error) {
	s.counters.iterators++
	iterator, err := s.KVStore.Iterator(start, end)
	if err != nil {
		return nil, err
	}
	return balanceCountingIterator{
		Iterator: iterator,
		counters: s.counters,
	}, nil
}

type balanceCountingIterator struct {
	corestore.Iterator
	counters *balanceStoreCounters
}

func (i balanceCountingIterator) Next() {
	i.counters.iteratorNexts++
	i.Iterator.Next()
}
