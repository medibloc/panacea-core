package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		DataValidators:     []*DataValidator{},
		NextPoolNumber:     uint64(1),
		Pools:              []*Pool{},
		Params:             DefaultParams(),
		NftContractAddress: nil,
		// this line is used by starport scaffolding # ibc/genesistype/default
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if gs.NftContractAddress != nil {
		if err := sdk.VerifyAddressFormat(gs.NftContractAddress); err != nil {
			return err
		}
	}
	// this line is used by starport scaffolding # ibc/genesistype/validate

	// this line is used by starport scaffolding # genesis/types/validate

	return nil
}
