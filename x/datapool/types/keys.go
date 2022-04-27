package types

import (
	"bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "datapool"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_datapool"

	// this line is used by starport scaffolding # ibc/keys/name
)

var (
	// KeyPrefixDataValidators defines key to store data validator
	KeyPrefixDataValidators = []byte{0x01}

	// KeyPoolNextNumber defines key to store next Pool ID to be used
	KeyPoolNextNumber = []byte{0x02}

	// KeyPrefixPools defines key to store Pools
	KeyPrefixPools = []byte{0x03}

	// KeyPrefixDataValidatorCerts defines key to store dataValidator certs
	KeyPrefixDataValidatorCerts = []byte{0x04}

	KeyPrefixRevenueDistribute = []byte{0x05}

	// KeyPrefixInstantRevenueDistribute defines key to distribute reward pool
	KeyPrefixInstantRevenueDistribute = []byte{0x06}

	KeyIndexSeparator = []byte{0x07}
)

func GetKeyPrefixDataValidator(dataValidatorAddr sdk.AccAddress) []byte {
	return append(KeyPrefixDataValidators, dataValidatorAddr.Bytes()...)
}

func GetKeyPrefixPools(poolID uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolID)...)
}

func GetKeyPrefixDataValidateCerts(poolID, round *uint64) []byte {
	prefixOfPoolID := append(KeyPrefixDataValidatorCerts, sdk.Uint64ToBigEndian(*poolID)...)
	if round != nil {
		return append(prefixOfPoolID, sdk.Uint64ToBigEndian(*round)...)
	}
	return prefixOfPoolID
}

func GetKeyPrefixDataValidateCert(poolID, round uint64, dataHash []byte) []byte {
	return append(GetKeyPrefixDataValidateCerts(&poolID, &round), dataHash...)
}

func GetKeyPrefixDataValidateCertByRound(poolID, round uint64) []byte {
	keyPoolAppended := append(KeyPrefixDataValidatorCerts, sdk.Uint64ToBigEndian(poolID)...)
	return append(keyPoolAppended, sdk.Uint64ToBigEndian(round)...)
}

func GetKeyPrefixSalesHistory(poolID, round uint64) []byte {
	return append(KeyPrefixRevenueDistribute, CombineKeys(sdk.Uint64ToBigEndian(poolID), sdk.Uint64ToBigEndian(round))...)
}

// CombineKeys function defines combines deal_id with data_hash.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
