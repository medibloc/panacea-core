package cli

const (
	FlagDealFile                    = "deal-file"
	DataVerificationCertificateFile = "data-verification-certificate-file"
)

type createDealInputs struct {
	DataSchema            []string `json:"data_schema"`
	Budget                string   `json:"budget"`
	MaxNumData            uint64   `json:"max_num_data"`
	TrustedDataValidators []string `json:"trusted_data_validators"`
}

type sellDataInputs struct {
	Cert   DataValidationCertification `json:"certificate"`
	Seller string                      `json:"seller"`
}

type DataValidationCertification struct {
	UnsignedCert    UnsignedDataValidationCertification `json:"unsigned_cert"`
	SignatureBase64 string                              `json:"signature_base64"`
}

type UnsignedDataValidationCertification struct {
	DealId                 uint64 `json:"deal_id"`
	DataHashBase64         string `json:"data_hash_base64"`
	EncryptedDataUrlBase64 string `json:"encrypted_data_url_base64"`
	DataValidatorAddress   string `json:"data_validator_address"`
	RequesterAddress       string `json:"requester_address"`
}
