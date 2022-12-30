package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "oracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_oracle"
)

var (
	// OraclesKey defines key to store oracle
	OraclesKey            = []byte{0x01}
	OracleRegistrationKey = []byte{0x02}
	OracleUpgradeInfoKey  = []byte{0x03}
	OracleUpgradeKey      = []byte{0x04}
	OracleUpgradeQueueKey = []byte{0x05}

	IndexSeparator = []byte{0xFF}
)

func GetOracleKey(address sdk.AccAddress) []byte {
	return append(OraclesKey, address...)
}

func GetOracleRegistrationKey(uniqueID string, address sdk.AccAddress) []byte {
	return append(OracleRegistrationKey, CombineKeys([]byte(uniqueID), address)...)
}

func GetOracleUpgradeKey(uniqueID string, address sdk.AccAddress) []byte {
	return append(OracleUpgradeKey, CombineKeys([]byte(uniqueID), address)...)
}

func GetOracleUpgradeQueueKey(uniqueID string, addr sdk.AccAddress) []byte {
	return append(OracleUpgradeQueueKey, CombineKeys([]byte(uniqueID), addr)...)
}

func GetOracleUpgradesKey(uniqueID string) []byte {
	return append(OracleUpgradeQueueKey, []byte(uniqueID)...)
}

func SplitOracleUpgradeQueueKey(key []byte) (string, sdk.AccAddress) {
	return string(key[1 : len(key)-21]), key[len(key)-20:]
}

func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, IndexSeparator)
}
