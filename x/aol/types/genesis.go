package types

type GenesisState struct {
	Owners  map[string]Owner  `json:"owners"`
	Topics  map[string]Topic  `json:"topics"`
	Writers map[string]Writer `json:"writers"`
	Records map[string]Record `json:"records"`
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func ValidateGenesis(data GenesisState) error {
	for bz := range data.Owners {
		var key GenesisOwnerKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	for bz := range data.Topics {
		var key GenesisTopicKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	for bz := range data.Writers {
		var key GenesisWriterKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	for bz := range data.Records {
		var key GenesisRecordKey
		err := key.Unmarshal(bz)
		if err != nil {
			return err
		}
	}
	return nil
}
