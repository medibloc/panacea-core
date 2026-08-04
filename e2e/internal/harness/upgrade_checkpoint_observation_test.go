package harness

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpgradeCheckpointObservationValidate(t *testing.T) {
	t.Parallel()
	valid := UpgradeCheckpointObservation{
		ObservedAt:    time.Now().UTC(),
		Node:          "fullnode-0",
		QueryBoundary: UpgradeCheckpointQueryBoundaryCometBFTRPC,
		Height:        77,
		BlockID:       "AABB",
		AppHash:       "CCDD",
	}
	require.NoError(t, valid.Validate())

	for _, test := range []struct {
		name   string
		mutate func(*UpgradeCheckpointObservation)
		want   string
	}{
		{name: "time", mutate: func(value *UpgradeCheckpointObservation) { value.ObservedAt = time.Time{} }, want: "observed_at"},
		{name: "node", mutate: func(value *UpgradeCheckpointObservation) { value.Node = "" }, want: "node"},
		{name: "boundary", mutate: func(value *UpgradeCheckpointObservation) { value.QueryBoundary = "cli" }, want: "query_boundary"},
		{name: "height", mutate: func(value *UpgradeCheckpointObservation) { value.Height = 0 }, want: "height"},
		{name: "block ID", mutate: func(value *UpgradeCheckpointObservation) { value.BlockID = "" }, want: "block_id"},
		{name: "app hash", mutate: func(value *UpgradeCheckpointObservation) { value.AppHash = "" }, want: "app_hash"},
	} {
		t.Run(test.name, func(t *testing.T) {
			candidate := valid
			test.mutate(&candidate)
			require.ErrorContains(t, candidate.Validate(), test.want)
		})
	}
}
