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
	// DealNextNumberKey defines key to store the next Deal ID to be used
	DealNextNumberKey = []byte{0x01}

	// DealKey defines key to store deals
	DealKey = []byte{0x02}

	// CertificateKey defines key to store certificate
	CertificateKey = []byte{0x03}

	KeyIndexSeparator = []byte{0xFF}
)

func GetDealKey(dealID uint64) []byte {
	return append(DealKey, sdk.Uint64ToBigEndian(dealID)...)
}

func GetCertificateKey(dealID uint64, dataHash string) []byte {
	return append(CertificateKey, CombineKeys(sdk.Uint64ToBigEndian(dealID), []byte(dataHash))...)
}

// CombineKeys function defines combines deal_id with data_hash.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
