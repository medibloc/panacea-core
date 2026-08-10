package e2e_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestQueryUpgradeDIDDeactivatedAtHeightUsesPublicNotFoundContract(t *testing.T) {
	t.Parallel()

	const (
		step   = "upgrade-post-deactivated-did"
		did    = "did:panacea:deactivated-fixture"
		height = int64(88)
	)
	var (
		gotStep       string
		gotDiagnostic string
		gotCommand    []string
	)
	query := func(_ context.Context, queryStep, expectedDiagnostic string, command ...string) error {
		gotStep = queryStep
		gotDiagnostic = expectedDiagnostic
		gotCommand = append([]string(nil), command...)
		return nil
	}

	evidence, err := queryUpgradeDIDDeactivatedAtHeight(
		context.Background(),
		query,
		step,
		did,
		height,
	)
	require.NoError(t, err)
	require.Equal(t, step, gotStep)
	require.Equal(t, "code = NotFound desc = DID deactivated", gotDiagnostic)
	require.Equal(t, []string{"did", "get-did", did, "--height", "88"}, gotCommand)
	require.Equal(t, upgradeDIDDeactivatedQueryEvidence{
		DID:                did,
		QueryHeight:        height,
		ExpectedGRPCCode:   "NotFound",
		ExpectedDiagnostic: "code = NotFound desc = DID deactivated",
		Observed:           true,
	}, evidence)
}

func TestQueryUpgradeDIDDeactivatedAtHeightRejectsInvalidOrMismatchedEvidence(t *testing.T) {
	t.Parallel()

	queryCalled := false
	query := func(_ context.Context, _, _ string, _ ...string) error {
		queryCalled = true
		return errors.New("DID not found")
	}
	_, err := queryUpgradeDIDDeactivatedAtHeight(
		context.Background(),
		query,
		"invalid-height",
		"did:panacea:fixture",
		0,
	)
	require.ErrorContains(t, err, "positive")
	require.False(t, queryCalled)

	_, err = queryUpgradeDIDDeactivatedAtHeight(
		context.Background(),
		query,
		"wrong-rejection",
		"did:panacea:fixture",
		89,
	)
	require.ErrorContains(t, err, "DID not found")
	require.True(t, queryCalled)
}

func TestValidateUpgradeDIDCheckpointContract(t *testing.T) {
	t.Parallel()
	recordedAt := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	observation := harness.UpgradeCheckpointObservation{
		ObservedAt:    recordedAt,
		Node:          "fullnode-0",
		QueryBoundary: harness.UpgradeCheckpointQueryBoundaryCometBFTRPC,
		Height:        77,
		BlockID:       "AABB",
		AppHash:       "CCDD",
	}
	require.NoError(t, validateUpgradeDIDCheckpointContract(recordedAt, 77, observation))

	for _, test := range []struct {
		name        string
		recordedAt  time.Time
		queryHeight int64
		observation harness.UpgradeCheckpointObservation
		want        string
	}{
		{name: "invalid observation", recordedAt: recordedAt, queryHeight: 77, observation: harness.UpgradeCheckpointObservation{}, want: "observation"},
		{name: "height mismatch", recordedAt: recordedAt, queryHeight: 78, observation: observation, want: "height"},
		{name: "time mismatch", recordedAt: recordedAt.Add(time.Second), queryHeight: 77, observation: observation, want: "observed_at"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateUpgradeDIDCheckpointContract(test.recordedAt, test.queryHeight, test.observation)
			require.ErrorContains(t, err, test.want)
		})
	}
}
