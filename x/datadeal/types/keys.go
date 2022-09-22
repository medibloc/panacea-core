package types

import (
	"bytes"

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
	KeyIndexSeparator = []byte{0xFF}

	DataSaleKey = []byte{0x02}
)

func GetDataSaleKey(verifiableCID []byte, dealID uint64) []byte {
	return append(DataSaleKey, CombineKeys(verifiableCID, sdk.Uint64ToBigEndian(dealID))...)
}

// CombineKeys function defines combines deal_id with data_hash.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
