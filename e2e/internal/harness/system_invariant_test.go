package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSystemEvidenceStepAcceptsArtifactSafePhases(t *testing.T) {
	t.Parallel()

	for _, step := range []string{"post-upgrade", "post-restart", "phase1"} {
		require.NoError(t, validateSystemEvidenceStep(step), step)
	}
}

func TestValidateSystemEvidenceStepRejectsPathsAndEmptyValues(t *testing.T) {
	t.Parallel()

	for _, step := range []string{"", "../escape", "post/restart", "Post-Upgrade", "post_upgrade"} {
		require.Error(t, validateSystemEvidenceStep(step), step)
	}
}
