package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDetachedCometStateSyncQueryRecord(t *testing.T) {
	tests := []struct {
		name       string
		height     int64
		wantArgs   []string
		historical bool
	}{
		{
			name:     "latest",
			height:   0,
			wantArgs: []string{"bank", "balances", "panacea1query"},
		},
		{
			name:       "historical",
			height:     30,
			wantArgs:   []string{"bank", "balances", "panacea1query", "--height", "30"},
			historical: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := []string{"bank", "balances", "panacea1query"}
			arguments, record := newDetachedCometStateSyncQueryRecord(
				"state-sync-query",
				"after-restart",
				"state-sync-node",
				command,
				test.height,
			)

			require.Equal(t, test.wantArgs, arguments)
			require.Equal(t, "state-sync-cli", record.Boundary)
			require.Equal(t, "state-sync-query", record.Step)
			require.Equal(t, test.height, record.Height)
			require.Equal(t, test.historical, record.HistoricalHeight)

			request, ok := record.Request.(map[string]any)
			require.True(t, ok)
			require.Equal(t, test.wantArgs, request["arguments"])
			require.Equal(t, test.height, request["height"])
			require.Equal(t, map[string]any{
				"node":  "state-sync-node",
				"phase": "after-restart",
			}, record.Metadata)

			command[0] = "mutated"
			require.Equal(t, "bank", arguments[0])
		})
	}
}
