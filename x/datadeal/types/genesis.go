package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Deals:          map[uint64]Deal{},
		DataCerts:      map[string]DataCert{},
		NextDealNumber: uint64(1),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for dealMapKey, deal := range gs.Deals {
		if dealMapKey != deal.DealId {
			return sdkerrors.Wrapf(ErrInvalidGenesisDeal, "dealID: %v", deal.DealId)
		}
	}

	for certMapKey, cert := range gs.DataCerts {
		key := string(GetKeyPrefixDataCert(cert.UnsignedCert.GetDealId(), cert.UnsignedCert.GetDataHash()))
		if certMapKey != key {
			return sdkerrors.Wrapf(ErrInvalidGenesisDataCert, "dealID: %v, dataHash: %s", cert.UnsignedCert.GetDealId(), string(cert.UnsignedCert.GetDataHash()))
		}
	}

	return nil
}
