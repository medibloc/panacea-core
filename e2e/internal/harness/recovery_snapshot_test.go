package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLocalSnapshotsFindsRequestedHeight(t *testing.T) {
	output := []byte("height: 18 format: 1 chunks: 3\nheight: 42 format: 1 chunks: 7\n")

	snapshots, err := ParseLocalSnapshots(output)
	require.NoError(t, err)
	require.Equal(t, []LocalSnapshot{
		{Height: 18, Format: 1, Chunks: 3},
		{Height: 42, Format: 1, Chunks: 7},
	}, snapshots)

	snapshot, err := FindLocalSnapshot(snapshots, 42)
	require.NoError(t, err)
	require.Equal(t, LocalSnapshot{Height: 42, Format: 1, Chunks: 7}, snapshot)
}
