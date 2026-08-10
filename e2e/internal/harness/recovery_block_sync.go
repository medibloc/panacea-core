package harness

import (
	"fmt"
	"strings"
)

// ValidateBlockSync proves that a newly added node retained the same exact
// CometBFT block and application state root as an existing reference node.
func ValidateBlockSync(reference, added RecoveryCheckpoint) error {
	for _, observed := range []struct {
		label      string
		checkpoint RecoveryCheckpoint
	}{
		{label: "reference", checkpoint: reference},
		{label: "added node", checkpoint: added},
	} {
		if observed.checkpoint.Height <= 0 {
			return fmt.Errorf("%s checkpoint has invalid height %d", observed.label, observed.checkpoint.Height)
		}
		if strings.TrimSpace(observed.checkpoint.BlockID) == "" {
			return fmt.Errorf("%s checkpoint is missing block ID", observed.label)
		}
		if strings.TrimSpace(observed.checkpoint.AppHash) == "" {
			return fmt.Errorf("%s checkpoint is missing app hash", observed.label)
		}
	}
	if reference.Height != added.Height {
		return fmt.Errorf("block-sync height differs: reference=%d added=%d", reference.Height, added.Height)
	}
	if !strings.EqualFold(reference.BlockID, added.BlockID) {
		return fmt.Errorf("block ID differs at height %d: reference=%s added=%s", reference.Height, reference.BlockID, added.BlockID)
	}
	if !strings.EqualFold(reference.AppHash, added.AppHash) {
		return fmt.Errorf("app hash differs at height %d: reference=%s added=%s", reference.Height, reference.AppHash, added.AppHash)
	}
	return nil
}
