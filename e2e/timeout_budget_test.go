package e2e_test

import (
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var runnerTimeoutAssignment = regexp.MustCompile(`(?m)^([A-Z0-9_]+)=\$\{([A-Z0-9_]+):-([^}\s]+)\}$`)

func TestUnsupportedDBTimeoutBudgetPreservesFailureCleanup(t *testing.T) {
	t.Parallel()

	// A one-node failed startup may spend 45 seconds collecting artifacts and
	// another 45 seconds removing labeled Docker resources. Keep an additional
	// 30 seconds for Interchain.Close and artifact finalization.
	require.GreaterOrEqual(t,
		unsupportedDBChildTimeout-unsupportedDBOperationTimeout,
		2*time.Minute,
	)
	require.GreaterOrEqual(t,
		unsupportedDBParentTimeout-unsupportedDBChildTimeout,
		time.Minute,
	)
}

func TestFailureProbeTimeoutBudgetPreservesFailureCleanup(t *testing.T) {
	t.Parallel()

	// A two-node run may spend 60 seconds collecting artifacts and another
	// 45 seconds removing Docker resources. The remaining margin covers the
	// interchain close and artifact finalization steps.
	require.GreaterOrEqual(t,
		failureProbeChildTimeout-failureProbeOperationTimeout,
		3*time.Minute,
	)
}

func TestStandaloneRunnerSuiteTimeoutsPreserveCleanupBudgets(t *testing.T) {
	t.Parallel()

	runner, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	timeouts := make(map[string]time.Duration)
	for _, match := range runnerTimeoutAssignment.FindAllStringSubmatch(string(runner), -1) {
		if match[1] != match[2] {
			continue
		}
		parsed, parseErr := time.ParseDuration(match[3])
		if parseErr == nil {
			timeouts[match[1]] = parsed
		}
	}

	// Smoke has one eight-minute two-node scenario. Restart selects a 15-minute
	// recovery scenario with 3-node and 2-node networks, then a ten-minute
	// snapshot scenario with a 3-node network. These lower bounds leave time for
	// artifact collection, Docker cleanup, interchain close, and finalization.
	require.GreaterOrEqual(t, timeouts["E2E_TEST_TIMEOUT"], 12*time.Minute)
	require.GreaterOrEqual(t, timeouts["E2E_RESTART_TIMEOUT"], 35*time.Minute)
}
