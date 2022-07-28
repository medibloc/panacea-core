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
	// KeyPrefixOracle defines key to store oracle
	KeyPrefixOracle                 = []byte{0x01}
	KeyPrefixOracleRegistration     = []byte{0x02}
	KeyPrefixOracleRegistrationVote = []byte{0x03}

	KeyIndexSeparator = []byte{0xFF}
)

func GetKeyPrefixOracle(address string) []byte {
	return append(KeyPrefixOracle, []byte(address)...)
}

func GetKeyPrefixOracleRegistration(address sdk.AccAddress) []byte {
	return append(KeyPrefixOracleRegistration, address...)
}

func GetKeyPrefixOracleRegistrationVote(uniqueID string, votingTargetAddress, voterAddress sdk.AccAddress) []byte {
	return append(KeyPrefixOracleRegistrationVote, CombineKeys([]byte(uniqueID), votingTargetAddress, voterAddress)...)
}

func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, KeyIndexSeparator)
}
