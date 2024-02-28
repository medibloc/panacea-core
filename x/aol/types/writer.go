package types

import (
	"cosmossdk.io/errors"
	"regexp"
)

const maxMonikerLength = 70

func (w Writer) Validate() error {
	if err := validateMoniker(w.Moniker); err != nil {
		return err
	}
	if err := validateDescription(w.Description); err != nil {
		return err
	}
	return nil
}

func validateMoniker(moniker string) error {
	if len(moniker) > maxMonikerLength {
		return errors.Wrapf(ErrMessageTooLarge, "moniker (%d > %d)", len(moniker), maxMonikerLength)
	}

	// can be an empty string
	if !regexp.MustCompile("^[A-Za-z0-9._-]*$").MatchString(moniker) {
		return errors.Wrapf(ErrInvalidMoniker, "moniker %s", moniker)
	}

	return nil
}
