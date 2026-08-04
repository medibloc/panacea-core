package harness

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadNodeRuntimeSampleJSONPreservesZeroValueObservations(t *testing.T) {
	t.Parallel()

	catchingUp := false
	peers := 0
	mempoolTransactions := 0
	mempoolBytes := int64(0)
	encoded, err := json.Marshal(LoadNodeRuntimeSample{
		CatchingUp:          &catchingUp,
		Peers:               &peers,
		MempoolTransactions: &mempoolTransactions,
		MempoolBytes:        &mempoolBytes,
	})
	require.NoError(t, err)

	var artifact struct {
		CatchingUp          *bool  `json:"catching_up"`
		Peers               *int   `json:"peers"`
		MempoolTransactions *int   `json:"mempool_transactions"`
		MempoolBytes        *int64 `json:"mempool_bytes"`
	}
	require.NoError(t, json.Unmarshal(encoded, &artifact))
	require.NotNil(t, artifact.CatchingUp)
	require.False(t, *artifact.CatchingUp)
	require.NotNil(t, artifact.Peers)
	require.Zero(t, *artifact.Peers)
	require.NotNil(t, artifact.MempoolTransactions)
	require.Zero(t, *artifact.MempoolTransactions)
	require.NotNil(t, artifact.MempoolBytes)
	require.Zero(t, *artifact.MempoolBytes)
}

func TestParsePrometheusGoroutines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want uint64
	}{
		{
			name: "standard Go collector",
			body: "# HELP go_goroutines Number of goroutines.\n# TYPE go_goroutines gauge\ngo_goroutines 42\n",
			want: 42,
		},
		{
			name: "Hashicorp runtime gauge",
			body: "runtime_num_goroutines{service=\"panacead\"} 17\n",
			want: 17,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := parsePrometheusGoroutines([]byte(test.body))
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestParsePrometheusGoroutinesRejectsMissingOrInvalidMetric(t *testing.T) {
	t.Parallel()

	_, err := parsePrometheusGoroutines([]byte("process_cpu_seconds_total 1\n"))
	require.ErrorContains(t, err, "goroutine")

	_, err = parsePrometheusGoroutines([]byte("go_goroutines not-a-number\n"))
	require.ErrorContains(t, err, "not-a-number")
}
