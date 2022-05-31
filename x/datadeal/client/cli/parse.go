package cli

const (
	FlagDealFile                    = "deal-file"
	DataVerificationCertificateFile = "data-cert-file"
)

type createDealInputs struct {
	DataSchema            []string `json:"data_schema"`
	Budget                string   `json:"budget"`
	MaxNumData            uint64   `json:"max_num_data"`
	TrustedDataValidators []string `json:"trusted_data_validators"`
}
