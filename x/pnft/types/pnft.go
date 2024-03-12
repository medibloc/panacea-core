package types

import (
	"fmt"
)

func (m *Pnft) ValidateBasic() error {
	if m.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	if m.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if m.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if m.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}

	if m.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}

	if m.CreatedAt.IsZero() {
		return fmt.Errorf("createdAt cannot be zero value")
	}
	return nil
}
