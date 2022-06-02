package cli

type CreatePoolInput struct {
	DataSchema         []string `json:"data_schema"`
	TargetNumData      uint64   `json:"target_num_data"`
	MaxNFTSupply       uint64   `json:"max_nft_supply"`
	NFTPrice           string   `json:"nft_price"`
	TrustedOracles     []string `json:"trusted_oracles"`
	TrustedDataIssuers []string `json:"trusted_data_issuers"`
}
