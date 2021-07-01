package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		return sdkerrors.Wrapf(ErrMessageTooLarge, "key (%d > %d)", len(key), maxRecordKeyLength)
	}
	return nil
}

func validateRecordValue(value []byte) error {
	if len(value) > maxRecordValueLength {
		return sdkerrors.Wrapf(ErrMessageTooLarge, "value (%d > %d)", len(value), maxRecordValueLength)
	}
	return nil
}
