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

	DataSaleKey              = []byte{0x03}
	DataVerificationVoteKey  = []byte{0x04}
	DataVerificationQueueKey = []byte{0x05}

	DataDeliveryVoteKey  = []byte{0x06}
	DataDeliveryQueueKey = []byte{0x07}

	DealQueueKey = []byte{0x08}

	KeyIndexSeparator = []byte{0xFF}
)

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

func GetDealKey(dealID uint64) []byte {
	return append(KeyPrefixDeals, sdk.Uint64ToBigEndian(dealID)...)
}

func GetDataSaleKey(dataHash string, dealID uint64) []byte {
	return append(DataSaleKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(dataHash))...)
}

func GetDataSalesKey(dealID uint64) []byte {
	return append(DataSaleKey, sdk.Uint64ToBigEndian(dealID)...)
}

func GetDataVerificationVoteKey(dataHash string, voterAddress sdk.AccAddress, dealID uint64) []byte {
	return append(DataVerificationVoteKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(dataHash), voterAddress)...)
}

func GetDataVerificationVotesKey(dataHash string, dealID uint64) []byte {
	return append(DataVerificationVoteKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(dataHash))...)
}

func GetDataVerificationQueueKey(dataHash string, dealID uint64, endTime time.Time) []byte {
	return append(DataVerificationQueueKey, CombineKeys(sdk.FormatTimeBytes(endTime), sdk.Uint64ToBigEndian(dealID), []byte(dataHash))...)
}

func GetDataVerificationQueueKeyByTimeKey(endTime time.Time) []byte {
	return append(DataVerificationQueueKey, sdk.FormatTimeBytes(endTime)...)
}

func GetDataDeliveryVoteKey(dealID uint64, dataHash string, voterAddress sdk.AccAddress) []byte {
	return append(DataDeliveryVoteKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(dataHash), voterAddress)...)
}

func GetDataDeliveryVotesKey(dealID uint64, dataHash string) []byte {
	return append(DataDeliveryVoteKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(dataHash))...)
}

func GetDataDeliveryQueueKey(dealID uint64, dataHash string, endTime time.Time) []byte {
	return append(DataDeliveryQueueKey, CombineKeys(sdk.FormatTimeBytes(endTime), sdk.Uint64ToBigEndian(dealID), []byte(dataHash))...)
}

func GetDataDeliveryQueueByTimeKey(endTime time.Time) []byte {
	return append(DataDeliveryQueueKey, sdk.FormatTimeBytes(endTime)...)
}

func SplitDataQueueKey(key []byte) (*time.Time, uint64, string, error) {
	votingEndTime, err := sdk.ParseTimeBytes(key[1 : 1+lenTime])
	if err != nil {
		return nil, 0, "", err
	}
	return &votingEndTime, sdk.BigEndianToUint64(key[1+lenTime+1 : 1+lenTime+1+8]), string(key[1+lenTime+1+8+1:]), nil
}

func GetDealQueueKey(dealID uint64, deactivationHeight int64) []byte {
	return append(DealQueueKey, CombineKeys(sdk.Uint64ToBigEndian(uint64(deactivationHeight)), sdk.Uint64ToBigEndian(dealID))...)
}

func GetDealQueueByHeight(deactivationHeight int64) []byte {
	return append(DealQueueKey, sdk.Uint64ToBigEndian(uint64(deactivationHeight))...)
}

func SplitDealQueueKey(key []byte) (int64, uint64) {
	return int64(sdk.BigEndianToUint64(key[1:10])), sdk.BigEndianToUint64(key[10:])
}

// CombineKeys function defines combines deal_id with data_hash.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
