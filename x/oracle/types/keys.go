package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	// KeyPrefixOracles defines key to store oracle
	KeyPrefixOracles = []byte{0x01}
)

func GetKeyPrefixOracle(oracleAddr sdk.AccAddress) []byte {
	return append(KeyPrefixOracles, oracleAddr.Bytes()...)
}
