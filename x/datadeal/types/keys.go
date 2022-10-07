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

	DataSaleKey = []byte{0x03}

	DataDeliveryVoteKey = []byte{0x06}

	DataDeliveryQueueKey = []byte{0x07}
)

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

func GetDealKey(dealID uint64) []byte {
	return append(KeyPrefixDeals, sdk.Uint64ToBigEndian(dealID)...)
}

func GetDataSaleKey(verifiableCID string, dealID uint64) []byte {
	return append(DataSaleKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(verifiableCID))...)
}

func GetDataDeliveryVoteKey(dealID uint64, verifiableCID string, voterAddress sdk.AccAddress) []byte {
	return append(DataDeliveryVoteKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(verifiableCID), voterAddress)...)
}

func GetDataDeliveryVotesKey(dealID uint64, verifiableCID string) []byte {
	return append(DataDeliveryVoteKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(verifiableCID))...)
}

func GetDataDeliveryQueueKey(dealID uint64, verifiableCID string, endTime time.Time) []byte {
	return append(DataDeliveryQueueKey, CombineKeys(sdk.FormatTimeBytes(endTime), sdk.Uint64ToBigEndian(dealID), []byte(verifiableCID))...)
}

func GetDataDeliveryQueueByTimeKey(endTime time.Time) []byte {
	return append(DataDeliveryQueueKey, sdk.FormatTimeBytes(endTime)...)
}

func SplitDataDeliveryQueueKey(key []byte) (uint64, string) {
	return sdk.BigEndianToUint64(key[1+lenTime+1 : 1+lenTime+1+8]), string(key[1+lenTime+1+8+1:])
}

// CombineKeys function defines combines deal_id with data_hash.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
