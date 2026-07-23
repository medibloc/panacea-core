package keeper

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	corestore "cosmossdk.io/core/store"
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

type balanceQueryFixture struct {
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
		server: NewStandardQueryServer(fixture.keeper),
		goCtx:  sdk.WrapSDKContext(fixture.ctx),
		request: &upstreamnft.QueryBalanceRequest{
			ClassId: created.ClassId,
			Owner:   owner,
		},
	}
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
