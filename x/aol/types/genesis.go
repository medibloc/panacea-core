package types

import (
	"fmt"
	// this line is used by starport scaffolding # ibc/genesistype/import
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # ibc/genesistype/default
		// this line is used by starport scaffolding # genesis/types/default
		OwnerList:  []*Owner{},
		RecordList: []*Record{},
		WriterList: []*Writer{},
		TopicList:  []*Topic{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # ibc/genesistype/validate

	// this line is used by starport scaffolding # genesis/types/validate
	// Check for duplicated ID in owner
	ownerIdMap := make(map[uint64]bool)

	for _, elem := range gs.OwnerList {
		if _, ok := ownerIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for owner")
		}
		ownerIdMap[elem.Id] = true
	}
	// Check for duplicated ID in record
	recordIdMap := make(map[uint64]bool)

	for _, elem := range gs.RecordList {
		if _, ok := recordIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for record")
		}
		recordIdMap[elem.Id] = true
	}
	// Check for duplicated ID in writer
	writerIdMap := make(map[uint64]bool)

	for _, elem := range gs.WriterList {
		if _, ok := writerIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for writer")
		}
		writerIdMap[elem.Id] = true
	}
	// Check for duplicated ID in topic
	topicIdMap := make(map[uint64]bool)

	for _, elem := range gs.TopicList {
		if _, ok := topicIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for topic")
		}
		topicIdMap[elem.Id] = true
	}

	return nil
}
