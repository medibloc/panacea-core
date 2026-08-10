package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateBlockSyncRequiresExactHistory(t *testing.T) {
	reference := RecoveryCheckpoint{
		Node:    "validator-0",
		Height:  73,
		BlockID: "AABB",
		AppHash: "CCDD",
	}
	added := reference
	added.Node = "fullnode-1"
	added.BlockID = "aabb"
	added.AppHash = "ccdd"

	require.NoError(t, ValidateBlockSync(reference, added))

	changedBlock := added
	changedBlock.BlockID = "EEFF"
	require.ErrorContains(t, ValidateBlockSync(reference, changedBlock), "block ID")

	changedApp := added
	changedApp.AppHash = "0011"
	require.ErrorContains(t, ValidateBlockSync(reference, changedApp), "app hash")

	changedHeight := added
	changedHeight.Height++
	require.ErrorContains(t, ValidateBlockSync(reference, changedHeight), "height")
}

func TestValidateBlockSyncRejectsIncompleteCheckpoint(t *testing.T) {
	reference := RecoveryCheckpoint{Node: "validator-0", Height: 73, BlockID: "AABB", AppHash: "CCDD"}
	added := RecoveryCheckpoint{Node: "fullnode-1", Height: 73, BlockID: "AABB"}

	require.ErrorContains(t, ValidateBlockSync(reference, added), "missing app hash")
}
