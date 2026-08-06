package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStandaloneRunnerAdvertisesCompleteInterface(t *testing.T) {
	runner, err := filepath.Abs("../scripts/e2e/run.sh")
	require.NoError(t, err)
	info, err := os.Stat(runner)
	require.NoError(t, err)
	require.NotZero(t, info.Mode().Perm()&0o111, "standalone runner must be executable")

	for _, helpArgument := range []string{"help", "--help"} {
		output, commandErr := exec.Command("sh", runner, helpArgument).CombinedOutput()
		require.NoErrorf(t, commandErr, "help output:\n%s", output)
		for _, command := range []string{
			"check", "check-clean", "build-current", "build-v2.2.1",
			"build-images", "build-test-binary", "build", "unit", "smoke",
			"v2.2.1", "compatibility", "negative", "restart", "consensus",
			"upgrade", "cosmovisor", "upgrade-deep", "upgrade-chaos", "state-sync",
			"config-compat", "ibc-upgrade", "network-faults", "release-builds",
			"release-hardening", "release-hardening-inner", "load", "all",
		} {
			require.Containsf(t, string(output), command, "%s must advertise %s", helpArgument, command)
		}
	}
}

func TestReleaseHardeningAggregateWatchdogCoversPathValidation(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	childPIDPath := filepath.Join(repoRoot, "path-validator-child.pid")
	writeExecutable(t, filepath.Join(repoRoot, "scripts", "e2e", "validate-paths.sh"), `#!/bin/sh
sleep 30 &
child_pid=$!
printf '%s\n' "$child_pid" >"$AGGREGATE_TEST_CHILD_PID_FILE"
wait "$child_pid"
`)

	runID := "p0p1-path-timeout"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=6",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=2",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"AGGREGATE_TEST_CHILD_PID_FILE="+childPIDPath,
	)
	started := time.Now()
	output, err := command.CombinedOutput()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		forceKillAggregatePIDFromFile(childPIDPath)
		t.Fatalf("path-validation watchdog exceeded deadline after %s: %v\n%s", elapsed, ctx.Err(), output)
	}
	requireAggregateExitCode(t, err, 124, output)
	require.Less(t, elapsed, 10*time.Second)
	bootstrapRoot := requireOneAggregateBootstrapArtifact(t, repoRoot)
	require.FileExists(t, filepath.Join(bootstrapRoot, "release", "overall-timeout.txt"))
	requireAggregateFailure(t, bootstrapRoot, "overall-timeout")
	requireAggregateProcessFromFileGone(t, childPIDPath)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateRunnerRecordsDirtySourceBeforeCallingRunner(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	runnerMarker := filepath.Join(repoRoot, "runner-called.txt")
	writeExecutable(t, filepath.Join(fakeBin, "git"), `#!/bin/sh
case "${1:-}" in
  rev-parse) printf '%s\n' 0123456789abcdef0123456789abcdef01234567 ;;
  status) printf '%s\n' '?? dirty-source-fixture' ;;
  diff) ;;
  *) exit 97 ;;
esac
`)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
printf '%s\n' called >"$AGGREGATE_TEST_RUNNER_MARKER"
exit 99
`)

	runID := "p0p1-dirty-source"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	command := exec.Command("sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"AGGREGATE_TEST_RUNNER_MARKER="+runnerMarker,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	requireAggregateExitCode(t, err, 2, output)

	require.NoFileExists(t, runnerMarker)
	requireAggregateFailure(t, aggregateRoot, "validate-clean-source")
	status := readAggregateFile(t, aggregateRoot, "release", "aggregate-status.txt")
	require.Contains(t, status, "result=failed\n")
	require.Contains(t, status, "exit_code=2\n")
	require.Contains(t, status, "stage=validate-clean-source\n")
	require.Contains(t, readAggregateFile(t, aggregateRoot, "release", "source-status.txt"), "dirty-source-fixture")
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateRunnerRecordsInvalidRootInBootstrapArtifact(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	invalidRoot := filepath.Join(string(filepath.Separator), "dev", "null", "panacea-p0p1-invalid")
	command := exec.Command("sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID=p0p1-invalid-root",
		"E2E_ROOT="+invalidRoot,
		"E2E_GOCACHE="+filepath.Join(invalidRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(invalidRoot, "go-mod"),
	)
	output, err := command.CombinedOutput()
	requireAggregateExitCode(t, err, 1, output)
	require.NoDirExists(t, invalidRoot)

	bootstrapRoot := requireOneAggregateBootstrapArtifact(t, repoRoot)
	requireAggregateFailure(t, bootstrapRoot, "validate-paths")
	status := readAggregateFile(t, bootstrapRoot, "release", "aggregate-status.txt")
	require.Contains(t, status, "result=failed\n")
	require.Contains(t, status, "stage=validate-paths\n")
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateRunnerFailsClosedWhenProcessEnumerationIsUnavailable(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	runnerMarker := filepath.Join(repoRoot, "runner-called.txt")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "ps"), `#!/bin/sh
exit 91
`)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
printf '%s\n' called >"$AGGREGATE_TEST_RUNNER_MARKER"
exit 9
`)

	runID := "p0p1-no-process-enumeration"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	command := exec.Command("sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=30",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=5",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"AGGREGATE_TEST_RUNNER_MARKER="+runnerMarker,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	requireAggregateExitCode(t, err, 91, output)
	require.NoFileExists(t, runnerMarker, "inner suites must not start without reliable process ownership checks")
	bootstrapRoot := requireOneAggregateBootstrapArtifact(t, repoRoot)
	requireAggregateFailure(t, bootstrapRoot, "validate-process-control")
	status := readAggregateFile(t, bootstrapRoot, "release", "aggregate-status.txt")
	require.Contains(t, status, "result=failed\n")
	require.Contains(t, status, "stage=validate-process-control\n")
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateRunnerTimeoutKillsOwnedDescendants(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	runnerPIDPath := filepath.Join(repoRoot, "term-resistant-runner.pid")
	childPIDPath := filepath.Join(repoRoot, "term-resistant-child.pid")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
printf '%s\n' "$$" >"$AGGREGATE_TEST_RUNNER_PID_FILE"
sh -c 'trap "" TERM; while :; do :; done' &
child_pid=$!
printf '%s\n' "$child_pid" >"$AGGREGATE_TEST_CHILD_PID_FILE"
trap '' TERM
while :; do :; done
`)

	runID := "p0p1-timeout"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=6",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=2",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"AGGREGATE_TEST_RUNNER_PID_FILE="+runnerPIDPath,
		"AGGREGATE_TEST_CHILD_PID_FILE="+childPIDPath,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	started := time.Now()
	output, err := command.CombinedOutput()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		forceKillAggregatePIDFromFile(runnerPIDPath)
		forceKillAggregatePIDFromFile(childPIDPath)
		t.Fatalf("aggregate runner exceeded test deadline after %s: %v\n%s", elapsed, ctx.Err(), output)
	}
	requireAggregateExitCode(t, err, 124, output)
	require.GreaterOrEqual(t, elapsed, time.Second)
	require.Less(t, elapsed, 10*time.Second)
	timeoutPath := filepath.Join(aggregateRoot, "release", "overall-timeout.txt")
	if _, statErr := os.Stat(timeoutPath); statErr != nil {
		matches, _ := filepath.Glob(filepath.Join(repoRoot, ".local", "e2e", "*", "release", "overall-timeout.txt"))
		t.Fatalf("overall timeout marker missing at %s; matches=%v output:\n%s", timeoutPath, matches, output)
	}
	cleanupTimeoutPath := filepath.Join(aggregateRoot, "release", "cleanup-timeout.txt")
	if _, statErr := os.Stat(cleanupTimeoutPath); statErr != nil {
		trace, _ := os.ReadFile(filepath.Join(aggregateRoot, "release", "aggregate-watchdog-trace.txt"))
		t.Fatalf("cleanup timeout marker missing; trace:\n%s\noutput:\n%s", trace, output)
	}
	requireAggregateFailure(t, aggregateRoot, "overall-timeout")
	requireAggregateProcessFromFileGone(t, runnerPIDPath)
	requireAggregateProcessFromFileGone(t, childPIDPath)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateOverallTimeoutInvalidatesGateBeforeForcedExit(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	runnerPIDPath := filepath.Join(repoRoot, "term-resistant-runner.pid")
	gateRecreatorPIDPath := filepath.Join(repoRoot, "gate-recreator.pid")
	gateRecreatedMarker := filepath.Join(repoRoot, "gate-recreated.txt")
	dockerPIDPath := filepath.Join(repoRoot, "term-resistant-docker.pid")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
mkdir -p "$E2E_ROOT/run-abcdef123456" "$E2E_ROOT/release"
printf '%s\n' '{"fixture":"successful-gate"}' >"$E2E_ROOT/release/gate-manifest.json"
printf '%s\n' "$$" >"$AGGREGATE_TEST_RUNNER_PID_FILE"
trap '' TERM
(
  trap '' TERM
  while [ -e "$E2E_ROOT/release/gate-manifest.json" ]; do :; done
  printf '%s\n' '{"fixture":"recreated-gate"}' >"$E2E_ROOT/release/gate-manifest.json"
  printf '%s\n' recreated >"$AGGREGATE_TEST_GATE_RECREATED_MARKER"
  while :; do :; done
) &
printf '%s\n' "$!" >"$AGGREGATE_TEST_GATE_RECREATOR_PID_FILE"
while :; do :; done
`)
	writeExecutable(t, filepath.Join(fakeBin, "docker"), `#!/bin/sh
printf '%s\n' "$$" >>"$AGGREGATE_TEST_DOCKER_PID_FILE"
trap '' TERM
while :; do :; done
`)

	runID := "p0p1-overall-forced-exit"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=7",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=2",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"DOCKER="+filepath.Join(fakeBin, "docker"),
		"AGGREGATE_TEST_RUNNER_PID_FILE="+runnerPIDPath,
		"AGGREGATE_TEST_GATE_RECREATOR_PID_FILE="+gateRecreatorPIDPath,
		"AGGREGATE_TEST_GATE_RECREATED_MARKER="+gateRecreatedMarker,
		"AGGREGATE_TEST_DOCKER_PID_FILE="+dockerPIDPath,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	started := time.Now()
	output, err := command.CombinedOutput()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		forceKillAggregatePIDFromFile(runnerPIDPath)
		forceKillAggregatePIDFromFile(gateRecreatorPIDPath)
		forceKillAggregatePIDsFromFile(dockerPIDPath)
		t.Fatalf("overall-timeout forced exit exceeded deadline after %s: %v\n%s", elapsed, ctx.Err(), output)
	}
	require.Error(t, err, "the aggregate runner must be hard-killed after its EXIT cleanup also stalls")
	require.Less(t, elapsed, 12*time.Second)
	requireAggregateFailure(t, aggregateRoot, "overall-timeout")
	require.FileExists(t, gateRecreatedMarker, "fixture must exercise gate recreation after initial invalidation")
	require.NoFileExists(t, filepath.Join(aggregateRoot, "release", "gate-manifest.json"))
	status := readAggregateFile(t, aggregateRoot, "release", "aggregate-status.txt")
	require.Contains(t, status, "result=failed\n")
	require.Contains(t, status, "stage=overall-timeout\n")
	requireAggregateProcessFromFileGone(t, runnerPIDPath)
	requireAggregateProcessFromFileGone(t, gateRecreatorPIDPath)
	requireAggregateProcessesFromFileGone(t, dockerPIDPath)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateInvalidWatchdogIdentityInvalidatesGateImmediately(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	runnerPIDPath := filepath.Join(repoRoot, "waiting-runner.pid")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
mkdir -p "$E2E_ROOT/release"
printf '%s\n' '{"fixture":"successful-gate"}' >"$E2E_ROOT/release/gate-manifest.json"
printf '%s\n' "$$" >"$AGGREGATE_TEST_RUNNER_PID_FILE"
while :; do sleep 1; done
`)

	runID := "p0p1-invalid-watchdog"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	command := exec.Command("sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=6",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=2",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"AGGREGATE_TEST_RUNNER_PID_FILE="+runnerPIDPath,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	require.NoError(t, command.Start())
	waitResult := make(chan error, 1)
	go func() { waitResult <- command.Wait() }()
	commandCompleted := false
	defer func() {
		if commandCompleted {
			return
		}
		forceKillAggregatePIDFromFile(runnerPIDPath)
		_ = command.Process.Kill()
		select {
		case <-waitResult:
		case <-time.After(3 * time.Second):
		}
	}()

	requireAggregateFileEventually(t, filepath.Join(aggregateRoot, "release", "gate-manifest.json"), 3*time.Second)
	watchdogPIDPath := requireOneAggregatePathEventually(t,
		filepath.Join(repoRoot, ".local", "e2e", ".aggregate-control-*", "watchdog-pid.txt"),
		3*time.Second,
	)
	require.NoError(t, os.WriteFile(watchdogPIDPath, []byte("999999999\n"), 0o600))
	requireAggregateFileEventually(t, filepath.Join(aggregateRoot, "release", "gate-failure.json"), 5*time.Second)
	requireAggregateFailure(t, aggregateRoot, "watchdog-identity-invalid")
	require.NoFileExists(t, filepath.Join(aggregateRoot, "release", "gate-manifest.json"),
		"the watchdog must invalidate success evidence before the runner enters EXIT cleanup")
	requireAggregatePathGoneEventually(t, filepath.Join(filepath.Dir(watchdogPIDPath), "artifact-root.lock"), time.Second)

	forceKillAggregatePIDFromFile(runnerPIDPath)
	select {
	case err := <-waitResult:
		commandCompleted = true
		requireAggregateExitCode(t, err, 124, nil)
	case <-time.After(5 * time.Second):
		t.Fatal("aggregate runner did not exit after the invalid-identity fixture was released")
	}
	requireAggregateProcessFromFileGone(t, runnerPIDPath)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateRunnerAllowsCooperativeCleanupBeforeHardKill(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	runnerPIDPath := filepath.Join(repoRoot, "cooperative-runner.pid")
	cleanupMarker := filepath.Join(repoRoot, "cooperative-cleanup.txt")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
printf '%s\n' "$$" >"$AGGREGATE_TEST_RUNNER_PID_FILE"
trap 'printf "%s\n" cleaned >"$AGGREGATE_TEST_CLEANUP_MARKER"; exit 143' TERM
while :; do sleep 1; done
`)

	runID := "p0p1-cooperative-timeout"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=6",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=3",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=1",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"AGGREGATE_TEST_RUNNER_PID_FILE="+runnerPIDPath,
		"AGGREGATE_TEST_CLEANUP_MARKER="+cleanupMarker,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		forceKillAggregatePIDFromFile(runnerPIDPath)
		t.Fatalf("cooperative aggregate timeout exceeded deadline: %v\n%s", ctx.Err(), output)
	}
	requireAggregateExitCode(t, err, 124, output)
	require.FileExists(t, cleanupMarker)
	require.NoFileExists(t, filepath.Join(aggregateRoot, "release", "cleanup-timeout.txt"))
	requireAggregateFailure(t, aggregateRoot, "overall-timeout")
	requireAggregateProcessFromFileGone(t, runnerPIDPath)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateCleanupFailsClosedWhenDockerLabelsRemain(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	dockerCalls := filepath.Join(repoRoot, "docker-calls.txt")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
mkdir -p "$E2E_ROOT/run-abcdef123456" "$E2E_ROOT/release"
printf '%s\n' '{"premature":true}' >"$E2E_ROOT/release/gate-manifest.json"
exit 0
`)
	writeExecutable(t, filepath.Join(fakeBin, "docker"), `#!/bin/sh
printf '%s\n' "$*" >>"$AGGREGATE_TEST_DOCKER_CALLS"
case "$*" in
  'ps -aq --filter label=ibc-test=run-abcdef123456') printf '%s\n' persistent-container ;;
  'volume ls -q --filter label=ibc-test=run-abcdef123456') printf '%s\n' persistent-volume ;;
  'network ls -q --filter label=ibc-test=run-abcdef123456') printf '%s\n' persistent-network ;;
esac
exit 0
`)

	runID := "p0p1-docker-cleanup"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	command := exec.Command("sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=30",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=5",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"DOCKER="+filepath.Join(fakeBin, "docker"),
		"AGGREGATE_TEST_DOCKER_CALLS="+dockerCalls,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	requireAggregateExitCode(t, err, 1, output)
	requireAggregateFailure(t, aggregateRoot, "aggregate-cleanup")
	require.NoFileExists(t, filepath.Join(aggregateRoot, "release", "gate-manifest.json"))
	calls := readAggregateFile(t, dockerCalls)
	require.GreaterOrEqual(t, strings.Count(calls, "ps -aq --filter label=ibc-test=run-abcdef123456"), 2)
	require.GreaterOrEqual(t, strings.Count(calls, "volume ls -q --filter label=ibc-test=run-abcdef123456"), 2)
	require.GreaterOrEqual(t, strings.Count(calls, "network ls -q --filter label=ibc-test=run-abcdef123456"), 2)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateCleanupTimeoutIsBoundedAndRecordsFailedStatus(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	dockerPIDPath := filepath.Join(repoRoot, "stuck-docker.pid")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
mkdir -p "$E2E_ROOT/run-abcdef123456"
exit 7
`)
	writeExecutable(t, filepath.Join(fakeBin, "docker"), `#!/bin/sh
printf '%s\n' "$$" >>"$AGGREGATE_TEST_DOCKER_PID_FILE"
trap '' TERM
while :; do :; done
`)

	runID := "p0p1-cleanup-timeout"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=12",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=2",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"DOCKER="+filepath.Join(fakeBin, "docker"),
		"AGGREGATE_TEST_DOCKER_PID_FILE="+dockerPIDPath,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	started := time.Now()
	output, err := command.CombinedOutput()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		forceKillAggregatePIDsFromFile(dockerPIDPath)
		t.Fatalf("aggregate cleanup watchdog exceeded deadline after %s: %v\n%s", elapsed, ctx.Err(), output)
	}
	require.Error(t, err, "stuck Docker cleanup must fail the aggregate")
	require.Less(t, elapsed, 9*time.Second)
	require.FileExists(t, filepath.Join(aggregateRoot, "release", "cleanup-timeout.txt"))
	requireAggregateFailure(t, aggregateRoot, "aggregate-cleanup-timeout")
	status := readAggregateFile(t, aggregateRoot, "release", "aggregate-status.txt")
	require.Contains(t, status, "result=failed\n")
	require.NotContains(t, status, "result=running\n")
	requireAggregateProcessesFromFileGone(t, dockerPIDPath)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateCleanupTimeoutInvalidatesGateAndResumesRemainingLabels(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	firstDockerCall := filepath.Join(repoRoot, "first-docker-call")
	dockerCalls := filepath.Join(repoRoot, "docker-calls.txt")
	dockerPIDPath := filepath.Join(repoRoot, "stuck-docker.pid")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
mkdir -p "$E2E_ROOT/run-111111111111" "$E2E_ROOT/run-222222222222" "$E2E_ROOT/release"
printf '%s\n' '{"fixture":"successful-gate"}' >"$E2E_ROOT/release/gate-manifest.json"
exit 0
`)
	writeExecutable(t, filepath.Join(fakeBin, "docker"), `#!/bin/sh
printf '%s\n' "$*" >>"$AGGREGATE_TEST_DOCKER_CALLS"
if mkdir "$AGGREGATE_TEST_FIRST_DOCKER_CALL" 2>/dev/null; then
  printf '%s\n' "$$" >"$AGGREGATE_TEST_DOCKER_PID_FILE"
  trap '' TERM
  while :; do :; done
fi
exit 0
`)

	runID := "p0p1-cleanup-resume"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=12",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=2",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=3",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"DOCKER="+filepath.Join(fakeBin, "docker"),
		"AGGREGATE_TEST_FIRST_DOCKER_CALL="+firstDockerCall,
		"AGGREGATE_TEST_DOCKER_CALLS="+dockerCalls,
		"AGGREGATE_TEST_DOCKER_PID_FILE="+dockerPIDPath,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	started := time.Now()
	output, err := command.CombinedOutput()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		forceKillAggregatePIDsFromFile(dockerPIDPath)
		t.Fatalf("aggregate cleanup continuation exceeded deadline after %s: %v\n%s", elapsed, ctx.Err(), output)
	}
	requireAggregateExitCode(t, err, 1, output)
	require.Less(t, elapsed, 9*time.Second)
	requireAggregateFailure(t, aggregateRoot, "aggregate-cleanup-timeout")
	require.NoFileExists(t, filepath.Join(aggregateRoot, "release", "gate-manifest.json"))
	status := readAggregateFile(t, aggregateRoot, "release", "aggregate-status.txt")
	require.Contains(t, status, "result=failed\n")
	require.Contains(t, status, "stage=aggregate-cleanup-timeout\n")
	trace := readAggregateFile(t, aggregateRoot, "release", "aggregate-watchdog-trace.txt")
	require.Contains(t, trace, "event=runner-cont-after-child-kill")
	calls := readAggregateFile(t, dockerCalls)
	require.Contains(t, calls, "ps -aq --filter label=ibc-test=run-222222222222")
	requireAggregateProcessesFromFileGone(t, dockerPIDPath)
	requireNoAggregateControlDirs(t, repoRoot)
}

func TestReleaseHardeningAggregateRunnerPassesCapturedCommitAndDeadlineToInnerRunner(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningAggregateRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	runnerArgs := filepath.Join(repoRoot, "runner-args.txt")
	writeExecutable(t, filepath.Join(fakeBin, "git"), cleanAggregateFakeGit)
	writeExecutable(t, filepath.Join(fakeBin, "runner"), `#!/bin/sh
{
  printf 'args=%s\n' "$*"
  printf 'E2E_CURRENT_SOURCE_COMMIT=%s\n' "${E2E_CURRENT_SOURCE_COMMIT:-}"
  printf 'COMMIT=%s\n' "${COMMIT:-}"
  printf 'E2E_RELEASE_AGGREGATE_WORK_DEADLINE_EPOCH=%s\n' "${E2E_RELEASE_AGGREGATE_WORK_DEADLINE_EPOCH:-}"
} >"$AGGREGATE_TEST_RUNNER_ARGS"
mkdir -p "$E2E_ROOT/release"
printf '%s\n' '{"fixture":true}' >"$E2E_ROOT/release/gate-manifest.json"
exit 0
`)

	runID := "p0p1-success"
	aggregateRoot := filepath.Join(repoRoot, ".local", "e2e", runID)
	command := exec.Command("sh", scriptPath)
	command.Env = append(aggregateRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_AGGREGATE=1",
		"E2E_RELEASE_HARDENING_RUN_ID="+runID,
		"E2E_ROOT="+aggregateRoot,
		"E2E_GOCACHE="+filepath.Join(aggregateRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(aggregateRoot, "go-mod"),
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=30",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=5",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"E2E_RUNNER="+filepath.Join(fakeBin, "runner"),
		"AGGREGATE_TEST_RUNNER_ARGS="+runnerArgs,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "output:\n%s", output)
	args := readAggregateFile(t, runnerArgs)
	require.Contains(t, args, "args=release-hardening-inner")
	require.Contains(t, args, "E2E_CURRENT_SOURCE_COMMIT=0123456789abcdef0123456789abcdef01234567")
	require.Contains(t, args, "COMMIT=0123456789abcdef0123456789abcdef01234567")
	require.Contains(t, args, "E2E_RELEASE_AGGREGATE_WORK_DEADLINE_EPOCH=")
	status := readAggregateFile(t, aggregateRoot, "release", "aggregate-status.txt")
	require.Contains(t, status, "result=passed\n")
	require.NoFileExists(t, filepath.Join(aggregateRoot, "release", "gate-failure.json"))
	require.FileExists(t, filepath.Join(aggregateRoot, "release", "aggregate.log"))
	requireNoAggregateControlDirs(t, repoRoot)
}

const cleanAggregateFakeGit = `#!/bin/sh
case "${1:-}" in
  rev-parse) printf '%s\n' 0123456789abcdef0123456789abcdef01234567 ;;
  status|diff) ;;
  *) exit 97 ;;
esac
`

func copyReleaseHardeningAggregateRunner(t *testing.T) (string, string) {
	t.Helper()
	contents, err := os.ReadFile("../scripts/e2e/release-hardening-aggregate.sh")
	require.NoError(t, err)
	pathValidator, err := os.ReadFile("../scripts/e2e/validate-paths.sh")
	require.NoError(t, err)
	repoRoot := t.TempDir()
	scriptDir := filepath.Join(repoRoot, "scripts", "e2e")
	require.NoError(t, os.MkdirAll(scriptDir, 0o700))
	scriptPath := filepath.Join(scriptDir, "release-hardening-aggregate.sh")
	require.NoError(t, os.WriteFile(scriptPath, contents, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(scriptDir, "validate-paths.sh"), pathValidator, 0o700))
	return repoRoot, scriptPath
}

func aggregateRunnerTestEnv() []string {
	prefixes := []string{
		"PANACEA_E2E_RELEASE_AGGREGATE=",
		"E2E_ROOT=",
		"E2E_GOCACHE=",
		"E2E_GOMODCACHE=",
		"E2E_RELEASE_HARDENING_RUN_ID=",
		"E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=",
		"E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=",
		"E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=",
		"E2E_RUNNER=",
		"DOCKER=",
		"AGGREGATE_TEST_",
	}
	filtered := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		keep := true
		for _, prefix := range prefixes {
			if strings.HasPrefix(item, prefix) {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func writeExecutable(t *testing.T, path, contents string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o700))
}

func requireAggregateExitCode(t *testing.T, err error, want int, output []byte) {
	t.Helper()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("command error = %v, want exit code %d\n%s", err, want, output)
	}
	require.Equalf(t, want, exitError.ExitCode(), "output:\n%s", output)
}

func requireAggregateFailure(t *testing.T, root, stage string) {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(root, "release", "gate-failure.json"))
	require.NoError(t, err)
	var failure struct {
		SchemaVersion string `json:"schema_version"`
		Stage         string `json:"stage"`
		Error         string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(contents, &failure))
	require.Equal(t, "2", failure.SchemaVersion)
	require.Equal(t, stage, failure.Stage)
	require.NotEmpty(t, failure.Error)
}

func readAggregateFile(t *testing.T, elements ...string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(elements...))
	require.NoError(t, err)
	return string(contents)
}

func requireOneAggregateBootstrapArtifact(t *testing.T, repoRoot string) string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(repoRoot, ".local", "e2e"))
	require.NoError(t, err)
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "p0p1-bootstrap-") {
			matches = append(matches, filepath.Join(repoRoot, ".local", "e2e", entry.Name()))
		}
	}
	require.Len(t, matches, 1)
	return matches[0]
}

func requireNoAggregateControlDirs(t *testing.T, repoRoot string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		matches, err := filepath.Glob(filepath.Join(repoRoot, ".local", "e2e", ".aggregate-control-*"))
		require.NoError(t, err)
		if len(matches) == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("aggregate control directories remain: %v", matches)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func requireAggregateFileEventually(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("aggregate file did not appear before deadline: %s", path)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func requireAggregatePathGoneEventually(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("aggregate path remained past deadline: %s", path)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func requireOneAggregatePathEventually(t *testing.T, pattern string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		matches, err := filepath.Glob(pattern)
		require.NoError(t, err)
		if len(matches) == 1 {
			return matches[0]
		}
		if time.Now().After(deadline) {
			t.Fatalf("wanted one aggregate path matching %s before deadline; matches=%v", pattern, matches)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func requireAggregateProcessFromFileGone(t *testing.T, path string) {
	t.Helper()
	pids := readAggregatePIDs(t, path)
	require.Len(t, pids, 1)
	requireAggregateProcessGone(t, pids[0])
}

func requireAggregateProcessesFromFileGone(t *testing.T, path string) {
	t.Helper()
	for _, pid := range readAggregatePIDs(t, path) {
		requireAggregateProcessGone(t, pid)
	}
}

func readAggregatePIDs(t *testing.T, path string) []int {
	t.Helper()
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Fields(string(contents))
	require.NotEmpty(t, lines)
	pids := make([]int, 0, len(lines))
	for _, line := range lines {
		pid, parseErr := strconv.Atoi(line)
		require.NoError(t, parseErr)
		pids = append(pids, pid)
	}
	return pids
}

func requireAggregateProcessGone(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		process, findErr := os.FindProcess(pid)
		if findErr != nil || process.Signal(syscall.Signal(0)) != nil {
			return
		}
		if time.Now().After(deadline) {
			_ = process.Signal(syscall.SIGKILL)
			t.Fatalf("aggregate runner left descendant PID %d alive", pid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func forceKillAggregatePIDFromFile(path string) {
	forceKillAggregatePIDsFromFile(path)
}

func forceKillAggregatePIDsFromFile(path string) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Fields(string(contents)) {
		pid, parseErr := strconv.Atoi(line)
		if parseErr != nil {
			continue
		}
		process, findErr := os.FindProcess(pid)
		if findErr == nil {
			_ = process.Signal(syscall.SIGKILL)
		}
	}
}
