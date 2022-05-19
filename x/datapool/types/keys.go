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

	// KeyPrefixNFTRedeemReceipts define key to store redeemed receipts
	KeyPrefixNFTRedeemReceipts = []byte{0x05}

	KeyPrefixSalesHistory = []byte{0x06}

	// KeyPrefixInstantRevenueDistribution defines key to distribution reward pool
	KeyPrefixInstantRevenueDistribution = []byte{0x07}

	KeyIndexSeparator = []byte{0xFF}
)

func GetKeyPrefixDataValidator(dataValidatorAddr sdk.AccAddress) []byte {
	return append(KeyPrefixDataValidators, dataValidatorAddr.Bytes()...)
}

func GetKeyPrefixPools(poolID uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolID)...)
}

func GetKeyPrefixDataValidateCerts(poolID, round uint64) []byte {
	return append(KeyPrefixDataValidatorCerts, CombineKeys(sdk.Uint64ToBigEndian(poolID), sdk.Uint64ToBigEndian(round))...)
}

func GetKeyPrefixDataValidateCert(poolID, round uint64, dataHash []byte) []byte {
	return CombineKeys(GetKeyPrefixDataValidateCerts(poolID, round), dataHash)
}

func GetKeyPrefixSalesHistories(poolID, round uint64) []byte {
	return append(KeyPrefixSalesHistory, CombineKeys(sdk.Uint64ToBigEndian(poolID), sdk.Uint64ToBigEndian(round))...)
}

func GetKeyPrefixSalesHistory(poolID, round uint64, seller string) []byte {
	return CombineKeys(GetKeyPrefixSalesHistories(poolID, round), []byte(seller))
}

func GetKeyPrefixNFTRedeemReceiptByPoolID(poolID uint64) []byte {
	return append(KeyPrefixNFTRedeemReceipts, sdk.Uint64ToBigEndian(poolID)...)
}

func GetKeyPrefixNFTRedeemReceipt(poolID, round, dataPassID uint64) []byte {
	return CombineKeys(GetKeyPrefixNFTRedeemReceiptByPoolID(poolID), sdk.Uint64ToBigEndian(round), sdk.Uint64ToBigEndian(dataPassID))
}

func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
