package types

import (
	"bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
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
	OraclesKey                  = []byte{0x01}
	OracleRegistrationKey       = []byte{0x02}
	OracleRegistrationVotesKey  = []byte{0x03}
	OracleRegistrationsQueueKey = []byte{0x04}

	IndexSeparator = []byte{0xFF}
)

func GetOracleKey(address sdk.AccAddress) []byte {
	return append(OraclesKey, address...)
}

func GetOracleRegistrationKey(uniqueID string, address sdk.AccAddress) []byte {
	return append(OracleRegistrationKey, CombineKeys([]byte(uniqueID), address)...)
}

func GetOracleRegistrationVotesKey(uniqueID string, votingTargetAddress sdk.AccAddress) []byte {
	return append(OracleRegistrationVotesKey, CombineKeys([]byte(uniqueID), votingTargetAddress)...)
}

func GetOracleRegistrationVoteKey(uniqueID string, votingTargetAddress, voterAddress sdk.AccAddress) []byte {
	return append(OracleRegistrationVotesKey, CombineKeys([]byte(uniqueID), votingTargetAddress, voterAddress)...)
}

func GetOracleRegistrationQueueKey(uniqueID string, addr sdk.AccAddress, endTime time.Time) []byte {
	return append(OracleRegistrationsQueueKey, CombineKeys(sdk.FormatTimeBytes(endTime), []byte(uniqueID), addr)...)
}

func GetOracleRegistrationVoteQueueByTimeKey(endTime time.Time) []byte {
	return append(OracleRegistrationsQueueKey, sdk.FormatTimeBytes(endTime)...)
}

func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, IndexSeparator)
}
