package cli

type createPoolInput struct {
	DataSchema            []string `json:"data_schema"`
	TargetNumData         uint64   `json:"target_num_data"`
	MaxNFTSupply          uint64   `json:"max_nft_supply"`
	NFTPrice              string   `json:"nft_price"`
	TrustedDataValidators []string `json:"trusted_data_validators"`
	TrustedDataIssuers    []string `json:"trusted_data_issuers"`
	Deposit               string   `json:"deposit"`
	DownloadPeriod        string   `json:"download_period"`
}
