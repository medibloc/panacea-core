package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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

	// KeyNFTContractAddress defines key to contract address
	KeyNFTContractAddress = []byte{0x04}

	// KeyPrefixDataValidatorCerts defines key to store dataValidator certs
	KeyPrefixDataValidatorCerts = []byte{0x05}
)

func GetKeyPrefixDataValidator(dataValidatorAddr sdk.AccAddress) []byte {
	return append(KeyPrefixDataValidators, dataValidatorAddr.Bytes()...)
}

func GetKeyPrefixPools(poolID uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolID)...)
}

func GetKeyPrefixDataValidateCert(poolID, round uint64, dataHash []byte) []byte {
	poolAppend := append(KeyPrefixDataValidatorCerts, sdk.Uint64ToBigEndian(poolID)...)
	roundAppend := append(poolAppend, sdk.Uint64ToBigEndian(round)...)
	return append(roundAppend, dataHash...)
}
