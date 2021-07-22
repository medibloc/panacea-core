package types

import "github.com/medibloc/panacea-core/v2/types/compkey"

// this line is used by starport scaffolding # ibc/genesistype/import

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

const GenesisKeySeparator = "/"

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Owners:  map[string]*Owner{},
		Topics:  map[string]*Topic{},
		Writers: map[string]*Writer{},
		Records: map[string]*Record{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for keyStr := range gs.Owners {
		var key OwnerCompositeKey
		if err := compkey.DecodeFromString(keyStr, GenesisKeySeparator, &key); err != nil {
			return err
		}
	}
	for keyStr, topic := range gs.Topics {
		var key TopicCompositeKey
		if err := compkey.DecodeFromString(keyStr, GenesisKeySeparator, &key); err != nil {
			return err
		}
		if err := topic.Validate(); err != nil {
			return err
		}
	}
	for keyStr, writer := range gs.Writers {
		var key WriterCompositeKey
		if err := compkey.DecodeFromString(keyStr, GenesisKeySeparator, &key); err != nil {
			return err
		}
		if err := writer.Validate(); err != nil {
			return err
		}
	}
	for keyStr, record := range gs.Records {
		var key RecordCompositeKey
		if err := compkey.DecodeFromString(keyStr, GenesisKeySeparator, &key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}
