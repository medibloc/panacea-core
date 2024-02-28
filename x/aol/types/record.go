package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	maxRecordKeyLength   = 70
	maxRecordValueLength = 5000
)

func (r Record) Validate() error {
	if err := validateRecordKey(r.Key); err != nil {
		return err
	}
	if err := validateRecordValue(r.Key); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(r.WriterAddress); err != nil {
		return err
	}
	return nil
}

func validateRecordKey(key []byte) error {
	if len(key) > maxRecordKeyLength {
		return errors.Wrapf(ErrMessageTooLarge, "key (%d > %d)", len(key), maxRecordKeyLength)
	}
	return nil
}

func validateRecordValue(value []byte) error {
	if len(value) > maxRecordValueLength {
		return errors.Wrapf(ErrMessageTooLarge, "value (%d > %d)", len(value), maxRecordValueLength)
	}
	return nil
}
