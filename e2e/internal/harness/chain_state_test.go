package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSameHistoryRequiresBlockIDAndAppHash(t *testing.T) {
	t.Parallel()

	matching := []BlockEvidence{
		{Node: "validator-0", Height: 9, BlockID: "AABB", AppHash: "CCDD"},
		{Node: "fullnode-0", Height: 9, BlockID: "aabb", AppHash: "ccdd"},
	}
	require.NoError(t, validateSameHistory(matching))

	wrongBlock := append([]BlockEvidence(nil), matching...)
	wrongBlock[1].BlockID = "FFFF"
	require.ErrorContains(t, validateSameHistory(wrongBlock), "block ID")

	wrongApp := append([]BlockEvidence(nil), matching...)
	wrongApp[1].AppHash = "EEEE"
	require.ErrorContains(t, validateSameHistory(wrongApp), "app hash")

	wrongHeight := append([]BlockEvidence(nil), matching...)
	wrongHeight[1].Height = 10
	require.ErrorContains(t, validateSameHistory(wrongHeight), "height")
}

func TestValidateSameHistoryRejectsInsufficientEvidence(t *testing.T) {
	t.Parallel()
	require.Error(t, validateSameHistory(nil))
	require.Error(t, validateSameHistory([]BlockEvidence{{Node: "validator-0"}}))
}
