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
	defaults := make(map[string]string)
	for _, match := range runnerTimeoutAssignment.FindAllStringSubmatch(string(runner), -1) {
		if match[1] != match[2] {
			continue
		}
		defaults[match[1]] = match[3]
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

	// The expanded normal P0 matrix owns a 30-minute context. The outer test
	// deadlines must leave bounded artifact and Docker cleanup time, including
	// when the same scenario runs against a release-architecture image.
	require.GreaterOrEqual(t, timeouts["E2E_UPGRADE_TIMEOUT"], 35*time.Minute)
	require.GreaterOrEqual(t, timeouts["E2E_UPGRADE_DEEP_TIMEOUT"], 50*time.Minute)
	require.GreaterOrEqual(t, timeouts["E2E_RELEASE_UPGRADE_TIMEOUT"], 35*time.Minute)

	// release-hardening now runs all functional suites before the independently
	// bounded six-hour release build. Keep the aggregate watchdog at twelve hours
	// in both the public runner and its directly executable wrapper.
	require.Equal(t, "43200", defaults["E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS"])
	aggregate, err := os.ReadFile("../scripts/e2e/release-hardening-aggregate.sh")
	require.NoError(t, err)
	require.Contains(t, string(aggregate),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=${E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS:-43200}")
	release, err := os.ReadFile("../scripts/e2e/release-hardening.sh")
	require.NoError(t, err)
	require.Contains(t, string(release),
		"E2E_RELEASE_UPGRADE_TIMEOUT=${E2E_RELEASE_UPGRADE_TIMEOUT:-35m}")
}
