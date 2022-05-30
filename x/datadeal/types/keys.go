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
	// KeyDealNextNumber defines key to store the next Deal ID to be used
	KeyDealNextNumber = []byte{0x01}

	// KeyPrefixDeals defines key to store deals
	KeyPrefixDeals = []byte{0x02}

	// KeyPrefixDataCertificateStore defines key to store data certificate
	KeyPrefixDataCertificateStore = []byte{0x03}

	KeyIndexSeparator = []byte{0xFF}
)

func GetKeyPrefixDeals(dealID uint64) []byte {
	return append(KeyPrefixDeals, sdk.Uint64ToBigEndian(dealID)...)
}

func GetKeyPrefixDataCertificate(dealID uint64, dataHash []byte) []byte {
	keys := CombineKeys(sdk.Uint64ToBigEndian(dealID), dataHash)
	return append(KeyPrefixDataCertificateStore, keys...)
}

// CombineKeys function defines combines deal_id with data_hash.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
