package cli

import (
	"bytes"
	"encoding/json"
)

const (
	FlagDealFile    = "deal-file"
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
	UnsignedCert UnsignedDataValidationCertification `json:"unsigned_cert"`
	Signature    string                              `json:"signature"`
}

type UnsignedDataValidationCertification struct {
	DealId               uint64 `json:"deal_id"`
	DataHash             string `json:"data_hash"`
	EncryptedDataUrl     string `json:"encrypted_data_url"`
	DataValidatorAddress string `json:"data_validator_address"`
	RequesterAddress     string `json:"requester_address"`
}

func (input *createDealInputs) UnmarshalJSON(data []byte) error {
	var createDeal createDealInputs
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&createDeal); err != nil {
		return nil
	}

	*input = createDeal
	return nil
}

func (input *sellDataInputs) UnmarshalJSON(data []byte) error {
	var sellData sellDataInputs
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&sellData); err != nil {
		return err
	}

	*input = sellData
	return nil
}
