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

func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, IndexSeparator)
}
