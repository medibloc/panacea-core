package keeper

import nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"

// classStoreKey returns the byte representation of the nft class key
// reference from x/nft/keeper/keys.go
func classStoreKey(classID string) []byte {
	key := make([]byte, len(nftkeeper.ClassKey)+len(classID))
	copy(key, nftkeeper.ClassKey)
	copy(key[len(nftkeeper.ClassKey):], classID)
	return key
}
