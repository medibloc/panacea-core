package keeper

import (
	"context"

	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/runtime"
)

type queryStoreCounters struct {
	nft    storeReadCounters
	policy storeReadCounters
}

func (c *queryStoreCounters) reset() {
	c.nft.reset()
	c.policy.reset()
}

type storeReadCounters struct {
	gets          uint64
	has           uint64
	iterators     uint64
	iteratorNexts uint64
}

func (c *storeReadCounters) reset() {
	c.gets = 0
	c.has = 0
	c.iterators = 0
	c.iteratorNexts = 0
}

func newQueryCountingKeeper(
	fixture keeperFixture,
	counters *queryStoreCounters,
) Keeper {
	return NewKeeper(
		fixture.cdc,
		queryCountingStoreService{
			delegate: runtime.NewKVStoreService(fixture.nftService),
			counters: &counters.nft,
		},
		queryCountingStoreService{
			delegate: runtime.NewKVStoreService(fixture.policyService),
			counters: &counters.policy,
		},
		fixture.accountKeeper,
		testBankKeeper{},
		fixture.moduleAccountAddresses,
	)
}

type queryCountingStoreService struct {
	delegate corestore.KVStoreService
	counters *storeReadCounters
}

func (s queryCountingStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return queryCountingStore{
		KVStore:  s.delegate.OpenKVStore(ctx),
		counters: s.counters,
	}
}

type queryCountingStore struct {
	corestore.KVStore
	counters *storeReadCounters
}

func (s queryCountingStore) Get(key []byte) ([]byte, error) {
	s.counters.gets++
	return s.KVStore.Get(key)
}

func (s queryCountingStore) Has(key []byte) (bool, error) {
	s.counters.has++
	return s.KVStore.Has(key)
}

func (s queryCountingStore) Iterator(start, end []byte) (corestore.Iterator, error) {
	s.counters.iterators++
	iterator, err := s.KVStore.Iterator(start, end)
	if err != nil {
		return nil, err
	}
	return queryCountingIterator{
		Iterator: iterator,
		counters: s.counters,
	}, nil
}

func (s queryCountingStore) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	s.counters.iterators++
	iterator, err := s.KVStore.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}
	return queryCountingIterator{
		Iterator: iterator,
		counters: s.counters,
	}, nil
}

type queryCountingIterator struct {
	corestore.Iterator
	counters *storeReadCounters
}

func (i queryCountingIterator) Next() {
	i.counters.iteratorNexts++
	i.Iterator.Next()
}
