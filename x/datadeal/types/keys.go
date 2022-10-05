package types

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "datadeal"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_datadeal"
)

var (
	// KeyDealNextNumber defines key to store the next Deal ID to be used
	KeyDealNextNumber = []byte{0x01}

	// KeyPrefixDeals defines key to store deals
	KeyPrefixDeals = []byte{0x02}

	KeyIndexSeparator = []byte{0xFF}

	DataSaleKey             = []byte{0x03}
	DataVerificationVoteKey = []byte{0x04}
	DataSaleQueueKey        = []byte{0x05}
)

func GetDealKey(dealID uint64) []byte {
	return append(KeyPrefixDeals, sdk.Uint64ToBigEndian(dealID)...)
}

func GetDataSaleKey(verifiableCID string, dealID uint64) []byte {
	return append(DataSaleKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(verifiableCID))...)
}

func GetDataSaleQueueKey(verifiableCID string, dealID uint64, endTime time.Time) []byte {
	return append(DataSaleQueueKey, CombineKeys(sdk.FormatTimeBytes(endTime), []byte(verifiableCID), sdk.Uint64ToBigEndian(dealID))...)
}

func GetDataVerificationVoteKey(verifiableCID string, voterAddress sdk.AccAddress, dealID uint64) []byte {
	return append(DataVerificationVoteKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(verifiableCID), voterAddress)...)
}

// CombineKeys function defines combines deal_id with data_hash.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
