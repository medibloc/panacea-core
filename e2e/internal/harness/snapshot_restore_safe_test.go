package harness

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewApplicationSnapshotRestorePlanScopesEveryPathToOneNodeHome(t *testing.T) {
	t.Parallel()

	plan, err := newApplicationSnapshotRestorePlan(
		"/var/cosmos-chain/panacea-run-123",
		42,
		1,
	)

	require.NoError(t, err)
	require.Equal(t, "/var/cosmos-chain/panacea-run-123/data/application.db", plan.ApplicationDB)
	require.Equal(t, "/var/cosmos-chain/panacea-run-123/data/application.db.snapshot-backup-42-1", plan.BackupDB)
	require.Equal(t, "/var/cosmos-chain/panacea-run-123/snapshot-42-1.tar.gz", plan.Archive)

	for _, unsafeHome := range []string{
		"/",
		"relative/home",
		"/var/cosmos-chain",
		"/var/cosmos-chain/../other",
		"/var/cosmos-chain/node/../other",
	} {
		_, err := newApplicationSnapshotRestorePlan(unsafeHome, 42, 1)
		require.Error(t, err, unsafeHome)
	}
}

func TestExecuteApplicationDBSwapRestoresOriginalAfterRestoreFailure(t *testing.T) {
	t.Parallel()

	var calls []string
	_, err := executeApplicationDBSwap(applicationDBSwapOperations{
		MoveOriginalAside: func() error {
			calls = append(calls, "move-original")
			return nil
		},
		RestoreSnapshot: func() error {
			calls = append(calls, "restore-snapshot")
			return errors.New("restore failed")
		},
		ValidateRestored: func() error {
			calls = append(calls, "validate-restored")
			return nil
		},
		RemoveRestored: func() error {
			calls = append(calls, "remove-restored")
			return nil
		},
		MoveOriginalBack: func() error {
			calls = append(calls, "move-original-back")
			return nil
		},
	})

	require.ErrorContains(t, err, "restore failed")
	require.Equal(t, []string{
		"move-original",
		"restore-snapshot",
		"remove-restored",
		"move-original-back",
	}, calls)
}

func TestSelectLocalSnapshotAtHeightDistinguishesMissingAndDuplicate(t *testing.T) {
	t.Parallel()

	snapshot, found, err := selectLocalSnapshotAtHeight([]LocalSnapshot{
		{Height: 10, Format: 1, Chunks: 2},
		{Height: 20, Format: 1, Chunks: 4},
	}, 20)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, LocalSnapshot{Height: 20, Format: 1, Chunks: 4}, snapshot)

	_, found, err = selectLocalSnapshotAtHeight([]LocalSnapshot{
		{Height: 10, Format: 1, Chunks: 2},
	}, 20)
	require.NoError(t, err)
	require.False(t, found)

	_, _, err = selectLocalSnapshotAtHeight([]LocalSnapshot{
		{Height: 20, Format: 1, Chunks: 4},
		{Height: 20, Format: 1, Chunks: 4},
	}, 20)
	require.ErrorContains(t, err, "multiple")
}
