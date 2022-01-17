package cli

import (
	"bytes"
	"encoding/json"
)

const (
	FlagDealFile = "deal-file"
)

type createDealInputs struct {
	DataSchema            []string `json:"data_schema"`
	Budget                string   `json:"budget"`
	MaxNumData            uint64   `json:"max_num_data"`
	TrustedDataValidators []string `json:"trusted_data_validators"`
}

type XCreateDealInputs createDealInputs

func (input *createDealInputs) UnmarshalJSON(data []byte) error {
	var createDeal XCreateDealInputs
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&createDeal); err != nil {
		return nil
	}

	*input = createDealInputs(createDeal)
	return nil
}
