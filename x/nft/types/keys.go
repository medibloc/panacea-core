package types

const (
	// ModuleName is the runtime name of the Panacea NFT module.
	ModuleName = "nft"

	// StoreKey stores the canonical Cosmos SDK NFT state.
	StoreKey = ModuleName

	// PolicyStoreKey stores Panacea-specific NFT policy and lifecycle state.
	// It must not share a prefix with StoreKey because the SDK uses store names
	// as database prefixes and rejects potentially colliding names.
	PolicyStoreKey = "panacea_nft"

	// PolicyCodespace identifies Panacea-specific NFT errors and intentionally
	// matches the policy store namespace.
	PolicyCodespace = PolicyStoreKey

	// RouterKey is the legacy message routing key.
	RouterKey = ModuleName

	// QuerierRoute is the legacy query routing key.
	QuerierRoute = ModuleName

	// BasicNFTDataTypeURL is the canonical Any type URL for BasicNFTData.
	BasicNFTDataTypeURL = "/panacea.nft.v1.BasicNFTData"

	// NFTDataInterfaceName is the protobuf interface implemented by accepted NFT metadata.
	NFTDataInterfaceName = "panacea.nft.v1.NFTData"
)
