package cli

import (
	"bytes"
	"encoding/json"
	flag "github.com/spf13/pflag"
)

const (
	FlagDealFile = "deal-json-file"
)

type createDealInputs struct {
	DataSchema            []string `json:"data_schema"`
	Budget                string   `json:"budget"`
	WantDataCount         uint64   `json:"want_data_count"`
	TrustedDataValidators []string `json:"trusted_data_validators"`
}

type XCreateDealInputs createDealInputs

type XCreateDealInputsExceptions struct {
	XCreateDealInputs
	Other *string
}

func (input *createDealInputs) UnmarshalJSON(data []byte) error {
	var createDealEX XCreateDealInputsExceptions
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&createDealEX); err != nil {
		return nil
	}

	*input = createDealInputs(createDealEX.XCreateDealInputs)
	return nil
}

func FlagSetCreateDeal() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagDealFile, "" ,"Deal json file path")
	return fs
}
