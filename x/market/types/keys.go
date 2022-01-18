package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "market"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_market"
)

var (
	// KeyDealNextNumber defines key to store the next Deal ID to be used
	KeyDealNextNumber = []byte{0x01}

	// KeyPrefixDeals defines key to store deals
	KeyPrefixDeals = []byte{0x02}

	// KeyPrefixDataStore defines key to store data
	KeyPrefixDataStore = []byte{0x03}

	KeyIndexSeparator = []byte{0x07}
)

func GetKeyPrefixDeals(dealId uint64) []byte {
	return append(KeyPrefixDeals, sdk.Uint64ToBigEndian(dealId)...)
}

func GetKeyPrefixData(dealId uint64, dataHash string) []byte {
	beDealId := sdk.Uint64ToBigEndian(dealId)
	bytes := []byte(dataHash)
	keys := combineKeys(beDealId, bytes)
	return append(KeyPrefixDataStore, keys...)
}

// combineKeys function defines combines deal_id with data_hash.
func combineKeys(keys ...[]byte) []byte {
	combined := []byte{}
	for i, key := range keys {
		combined = append(combined, key...)
		if i < len(keys)-1 { // not last item
			combined = append(combined, KeyIndexSeparator...)
		}
	}
	return combined
}
