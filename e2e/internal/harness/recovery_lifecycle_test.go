package harness

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunStoppedNodeOperationRestartsAfterOperationFailure(t *testing.T) {
	var calls []string
	err := runRecoveryStoppedOperation(
		func() error {
			calls = append(calls, "stop")
			return nil
		},
		func() error {
			calls = append(calls, "operation")
			return errors.New("restore failed")
		},
		func() error {
			calls = append(calls, "start")
			return nil
		},
	)

	require.ErrorContains(t, err, "restore failed")
	require.Equal(t, []string{"stop", "operation", "start"}, calls)
}

func TestClassifyWALReplayEvidence(t *testing.T) {
	t.Parallel()

	replayLog := []byte("Catchup by replaying consensus messages height=17\nReplay: Done module=consensus")
	require.NoError(t, classifyWALReplayEvidence(replayLog))
	require.ErrorContains(t, classifyWALReplayEvidence([]byte("Starting baseWAL service")), "consensus replay start")
	require.ErrorContains(t, classifyWALReplayEvidence([]byte("Catchup by replaying consensus messages")), "consensus replay completion")
}
