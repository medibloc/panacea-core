package e2e_test

import (
	"context"
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

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestReleaseHardeningRunnerDoesNotDowngradeRequiredEvidenceToSkips(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/release-hardening.sh")
	require.NoError(t, err)
	script := string(contents)

	for _, contract := range []string{
		`"schema_version": "` + harness.ReleaseHardeningManifestSchemaVersion + `"`,
		`"source_clean": true`,
		`host-image-identity.json`,
		`E2E_FUNCTIONAL_CURRENT_IMAGE`,
		`E2E_FUNCTIONAL_OLD_IMAGE`,
		`source-status-final.txt`,
		`PANACEA_E2E_RELEASE_HARDENING`,
		`PANACEA_E2E_RELEASE_MULTIARCH_UPGRADE`,
		`E2E_RELEASE_TOTAL_TIMEOUT_SECONDS`,
		`E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS`,
		`E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS`,
		`cleanup-timeout.txt`,
		`release_is_descendant`,
		`release_signal_runner_children`,
		`validate-process-control`,
		`linux/amd64 linux/arm64`,
		`buildx create`,
		`--driver docker-container`,
		`--network=none`,
		`--no-cache`,
		`--metadata-file`,
		`GOPROXY=off`,
		`GOSUMDB=off`,
		`builder-cache-record-ids-before-build.txt`,
		`network disconnect`,
		`warm-offline-buildkit-contract.txt`,
		`DOCKER_HOST`,
		`dependencies-current.jsonl`,
		`dependencies-v2.2.1.jsonl`,
		`dependencies-e2e.jsonl`,
		`version --long`,
		`binary-sha256.txt`,
		`TestV221ToCurrentMultiValidatorUpgrade`,
		`TestValidateReleaseHardeningArtifact`,
	} {
		require.Contains(t, script, contract)
	}
	require.NotContains(t, script, "buildx create --use")
	require.NotContains(t, strings.ToLower(script), ":latest")
}

func TestReleaseHardeningRunnerPinsInputsToConfiguredGoVersion(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/release-hardening.sh")
	require.NoError(t, err)
	script := string(contents)

	require.Contains(t, script, `index($2, "golang:" version "-") == 1`)
	require.Contains(t, script, `grep -Fqx "go $E2E_GO_VERSION"`)
	require.NotContains(t, strings.ReplaceAll(script, `\`, ""), "1.23.12")
}

func TestReleaseHardeningMultiarchUpgradeUsesBuiltImageTags(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/release-hardening.sh")
	require.NoError(t, err)
	script := string(contents)

	require.Contains(t, script, `PANACEA_E2E_IMAGE_VERSION="$run_id"`)
	require.Contains(t, script, `PANACEA_E2E_V221_IMAGE_VERSION="$run_id"`)
	require.NotContains(t, script, `PANACEA_E2E_IMAGE_VERSION="$image_version"`)
	require.NotContains(t, script, `PANACEA_E2E_V221_IMAGE_VERSION="$image_version"`)
}

func TestReleaseHardeningMultiarchUpgradeRunsPrecompiledTestBinaryDirectly(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/release-hardening.sh")
	require.NoError(t, err)
	script := string(contents)

	require.Contains(t, script, `test -c -o "$release_test_binary" .`)
	require.Contains(t, script, `"$release_test_binary" -test.timeout "$E2E_RELEASE_UPGRADE_TIMEOUT"`)
	require.NotContains(t, script, `"$E2E_GO_BINARY" test -timeout "$E2E_RELEASE_UPGRADE_TIMEOUT"`)
}

func TestRootAutomationDoesNotInvokeStandaloneE2E(t *testing.T) {
	for _, path := range []string{
		"../Makefile",
		"../.github/workflows/ci.yml",
		"../.github/workflows/docker-publish.yml",
	} {
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		automation := strings.ToLower(string(contents))
		for _, forbidden := range []string{
			"scripts/e2e/run.sh",
			"test-e2e",
			"panacea_e2e",
			"release/gate-manifest.json",
		} {
			require.NotContainsf(t, automation, forbidden, "%s must leave E2E opt-in", path)
		}
	}
}

func TestReleaseHardeningRunnerRecordsOptInFailure(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningRunner(t)
	command := exec.Command("sh", scriptPath)
	command.Env = releaseRunnerTestEnv()
	output, err := command.CombinedOutput()
	requireExitCode(t, err, 2, output)

	artifactDir := requireOneReleaseRunnerArtifact(t, repoRoot)
	require.FileExists(t, filepath.Join(artifactDir, "failure.txt"))
	status, err := os.ReadFile(filepath.Join(artifactDir, "status.txt"))
	require.NoError(t, err)
	require.Contains(t, string(status), "result=failed\n")
	require.Contains(t, string(status), "exit_code=2\n")
	require.Contains(t, string(status), "stage=validate-opt-in\n")
	requireNoReleaseRunnerWorkDir(t, repoRoot)
}

func TestReleaseHardeningRunnerUsesValidatedCustomRootAndLeavesNoWatchdogSleep(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningRunner(t)

	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(fakeBin, "git"),
		[]byte(`#!/bin/sh
case "${1:-}" in
  rev-parse)
    "$RELEASE_TEST_REAL_SLEEP" 1
    printf '%s\n' 0123456789abcdef0123456789abcdef01234567
    ;;
  status)
    printf '%s\n' '?? injected-dirty-file'
    ;;
  diff) ;;
  *) exit 97 ;;
esac
`),
		0o700,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(fakeBin, "sleep"),
		[]byte(`#!/bin/sh
printf '%s\n' "$$" >"$RELEASE_TEST_SLEEP_PID_FILE"
exec "$RELEASE_TEST_REAL_SLEEP" "$@"
`),
		0o700,
	))
	realSleep, err := exec.LookPath("sleep")
	require.NoError(t, err)
	canonicalRepoRoot, err := filepath.EvalSymlinks(repoRoot)
	require.NoError(t, err)
	customRoot := filepath.Join(canonicalRepoRoot, ".local", "e2e", "p0p1-custom")
	watchdogSleepPIDPath := filepath.Join(repoRoot, "watchdog-sleep.pid")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(releaseRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_HARDENING=1",
		"E2E_ROOT="+customRoot,
		"E2E_GOCACHE="+filepath.Join(customRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(customRoot, "go-mod"),
		"E2E_RELEASE_TOTAL_TIMEOUT_SECONDS=30",
		"E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS=5",
		"E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"RELEASE_TEST_REAL_SLEEP="+realSleep,
		"RELEASE_TEST_SLEEP_PID_FILE="+watchdogSleepPIDPath,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		forceKillPIDFromFile(t, watchdogSleepPIDPath)
		t.Fatalf("release runner custom-root test exceeded deadline: %v\n%s", ctx.Err(), output)
	}
	requireExitCode(t, err, 2, output)

	artifactDir := requireOneReleaseRunnerArtifactAtRoot(t, customRoot)
	canonicalArtifactDir, canonicalErr := filepath.EvalSymlinks(artifactDir)
	require.NoError(t, canonicalErr)
	status, readErr := os.ReadFile(filepath.Join(artifactDir, "status.txt"))
	require.NoError(t, readErr)
	require.Contains(t, string(status), "stage=validate-clean-source\n")
	require.Contains(t, string(status), "artifact_dir="+canonicalArtifactDir+"\n")
	requireNoReleaseRunnerWorkDirAtRoot(t, customRoot)
	requireProcessFromFileGone(t, watchdogSleepPIDPath)

	bootstrapMatches, globErr := filepath.Glob(filepath.Join(repoRoot, ".local", "e2e", "release-*"))
	require.NoError(t, globErr)
	require.Empty(t, bootstrapMatches, "validated custom-root run must not leave bootstrap release artifacts")
}

func TestReleaseHardeningRunnerRejectsInvalidRootBeforeCreatingIt(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningRunner(t)
	canonicalRepoRoot, err := filepath.EvalSymlinks(repoRoot)
	require.NoError(t, err)
	invalidRoot := filepath.Join(string(filepath.Separator), "dev", "null", "panacea-e2e-root")

	command := exec.Command("sh", scriptPath)
	command.Env = append(releaseRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_HARDENING=1",
		"E2E_ROOT="+invalidRoot,
		"E2E_GOCACHE="+filepath.Join(invalidRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(invalidRoot, "go-mod"),
	)
	output, err := command.CombinedOutput()
	requireExitCode(t, err, 1, output)
	require.NoDirExists(t, invalidRoot)

	artifactDir := requireOneReleaseRunnerArtifact(t, canonicalRepoRoot)
	status, readErr := os.ReadFile(filepath.Join(artifactDir, "status.txt"))
	require.NoError(t, readErr)
	require.Contains(t, string(status), "stage=validate-paths\n")
	require.NoFileExists(t, filepath.Join(artifactDir, "watchdog-pid.txt"))
	requireNoReleaseRunnerWorkDir(t, canonicalRepoRoot)
}

func TestReleaseHardeningRunnerFailsClosedWhenProcessEnumerationIsUnavailable(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningRunner(t)
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(fakeBin, "ps"),
		[]byte("#!/bin/sh\nexit 91\n"),
		0o700,
	))
	gitMarker := filepath.Join(repoRoot, "git-called.txt")
	require.NoError(t, os.WriteFile(
		filepath.Join(fakeBin, "git"),
		[]byte("#!/bin/sh\nprintf '%s\\n' called >\"$RELEASE_TEST_GIT_MARKER\"\nexit 97\n"),
		0o700,
	))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.WaitDelay = time.Second
	command.Env = append(releaseRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_HARDENING=1",
		"RELEASE_TEST_GIT_MARKER="+gitMarker,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("release runner process-control preflight exceeded deadline: %v\n%s", ctx.Err(), output)
	}
	requireExitCode(t, err, 91, output)
	require.NoFileExists(t, gitMarker, "release work must not start without process enumeration")

	artifactDir := requireOneReleaseRunnerArtifact(t, repoRoot)
	failure, readErr := os.ReadFile(filepath.Join(artifactDir, "failure.txt"))
	require.NoError(t, readErr)
	require.Contains(t, string(failure), "result=failed\n")
	require.Contains(t, string(failure), "exit_code=91\n")
	require.Contains(t, string(failure), "stage=validate-process-control\n")
	status, readErr := os.ReadFile(filepath.Join(artifactDir, "status.txt"))
	require.NoError(t, readErr)
	require.Contains(t, string(status), "result=failed\n")
	require.Contains(t, string(status), "exit_code=91\n")
	require.Contains(t, string(status), "stage=validate-process-control\n")
	for _, name := range []string{
		"watchdog-pid.txt",
		"watchdog-timer-pid.txt",
		"watchdog-trace.txt",
		"overall-timeout.txt",
		"cleanup-timeout.txt",
	} {
		require.NoFileExists(t, filepath.Join(artifactDir, name))
	}
	requireNoReleaseRunnerWorkDir(t, repoRoot)
}

func TestReleaseHardeningRunnerTimeoutKillsOwnedDescendantsAndRecordsCleanupTimeout(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningRunner(t)
	scriptDir := filepath.Dir(scriptPath)
	require.NoError(t, os.WriteFile(
		filepath.Join(scriptDir, "validate-paths.sh"),
		[]byte("#!/bin/sh\nexit 0\n"),
		0o700,
	))

	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	childPIDPath := filepath.Join(repoRoot, "term-resistant-child.pid")
	fakeGit := `#!/bin/sh
if [ "${1:-}" = "rev-parse" ]; then
  sh -c 'trap "" TERM; while :; do :; done' &
  child_pid=$!
  printf '%s\n' "$child_pid" >"$RELEASE_TEST_CHILD_PID_FILE"
  trap '' TERM
  while :; do :; done
fi
exit 97
`
	require.NoError(t, os.WriteFile(filepath.Join(fakeBin, "git"), []byte(fakeGit), 0o700))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(releaseRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_HARDENING=1",
		"E2E_RELEASE_TOTAL_TIMEOUT_SECONDS=6",
		"E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS=2",
		"E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS=2",
		"RELEASE_TEST_CHILD_PID_FILE="+childPIDPath,
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	started := time.Now()
	output, err := command.CombinedOutput()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		forceKillPIDFromFile(t, childPIDPath)
		t.Fatalf("release runner exceeded test deadline after %s: %v\n%s", elapsed, ctx.Err(), output)
	}
	artifactDir := requireOneReleaseRunnerArtifact(t, repoRoot)
	trace, traceErr := os.ReadFile(filepath.Join(artifactDir, "watchdog-trace.txt"))
	require.NoError(t, traceErr)
	require.Error(t, err, "forced cleanup timeout must fail the runner")
	require.GreaterOrEqual(t, elapsed, time.Second)
	require.Less(t, elapsed, 10*time.Second)

	require.FileExists(t, filepath.Join(artifactDir, "overall-timeout.txt"))
	require.FileExistsf(t, filepath.Join(artifactDir, "cleanup-timeout.txt"), "watchdog trace:\n%s", trace)
	failure, err := os.ReadFile(filepath.Join(artifactDir, "failure.txt"))
	require.NoError(t, err)
	require.Contains(t, string(failure), "result=failed\n")
	require.Contains(t, string(failure), "exit_code=124\n")
	status, err := os.ReadFile(filepath.Join(artifactDir, "status.txt"))
	require.NoError(t, err)
	require.Contains(t, string(status), "result=failed\n")
	require.Contains(t, string(status), "exit_code=124\n")
	requireNoReleaseRunnerWorkDir(t, repoRoot)
	requireProcessFromFileGone(t, childPIDPath)
}

func TestReleaseHardeningRunnerCooperativeTimeoutReturns124(t *testing.T) {
	repoRoot, scriptPath := copyReleaseHardeningRunner(t)
	scriptDir := filepath.Dir(scriptPath)
	require.NoError(t, os.WriteFile(
		filepath.Join(scriptDir, "validate-paths.sh"),
		[]byte("#!/bin/sh\nexit 0\n"),
		0o700,
	))
	fakeBin := filepath.Join(repoRoot, "fake-bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(fakeBin, "git"),
		[]byte("#!/bin/sh\nsleep 30\n"),
		0o700,
	))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", scriptPath)
	command.Env = append(releaseRunnerTestEnv(),
		"PANACEA_E2E_RELEASE_HARDENING=1",
		"E2E_RELEASE_TOTAL_TIMEOUT_SECONDS=6",
		"E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS=4",
		"E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS=1",
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("cooperative release runner timeout exceeded test deadline: %v\n%s", ctx.Err(), output)
	}
	requireExitCode(t, err, 124, output)
	artifactDir := requireOneReleaseRunnerArtifact(t, repoRoot)
	require.FileExists(t, filepath.Join(artifactDir, "overall-timeout.txt"))
	require.NoFileExists(t, filepath.Join(artifactDir, "cleanup-timeout.txt"))
	status, readErr := os.ReadFile(filepath.Join(artifactDir, "status.txt"))
	require.NoError(t, readErr)
	require.Contains(t, string(status), "stage=overall-timeout\n")
	requireNoReleaseRunnerWorkDir(t, repoRoot)
}

func copyReleaseHardeningRunner(t *testing.T) (string, string) {
	t.Helper()
	contents, err := os.ReadFile("../scripts/e2e/release-hardening.sh")
	require.NoError(t, err)
	pathValidator, err := os.ReadFile("../scripts/e2e/validate-paths.sh")
	require.NoError(t, err)
	repoRoot := t.TempDir()
	scriptDir := filepath.Join(repoRoot, "scripts", "e2e")
	require.NoError(t, os.MkdirAll(scriptDir, 0o700))
	scriptPath := filepath.Join(scriptDir, "release-hardening.sh")
	require.NoError(t, os.WriteFile(scriptPath, contents, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(scriptDir, "validate-paths.sh"), pathValidator, 0o700))
	return repoRoot, scriptPath
}

func releaseRunnerTestEnv() []string {
	prefixes := []string{
		"PANACEA_E2E_RELEASE_HARDENING=",
		"E2E_ROOT=",
		"E2E_GOCACHE=",
		"E2E_GOMODCACHE=",
		"E2E_RELEASE_TOTAL_TIMEOUT_SECONDS=",
		"E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS=",
		"E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS=",
		"RELEASE_TEST_CHILD_PID_FILE=",
		"RELEASE_TEST_GIT_MARKER=",
		"RELEASE_TEST_REAL_SLEEP=",
		"RELEASE_TEST_SLEEP_PID_FILE=",
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

func requireOneReleaseRunnerArtifact(t *testing.T, repoRoot string) string {
	t.Helper()
	return requireOneReleaseRunnerArtifactAtRoot(t, filepath.Join(repoRoot, ".local", "e2e"))
}

func requireOneReleaseRunnerArtifactAtRoot(t *testing.T, artifactRoot string) string {
	t.Helper()
	entries, err := os.ReadDir(artifactRoot)
	require.NoError(t, err)
	var artifactDirs []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "release-") && !strings.HasSuffix(entry.Name(), "-work") {
			artifactDirs = append(artifactDirs, filepath.Join(artifactRoot, entry.Name()))
		}
	}
	require.Len(t, artifactDirs, 1)
	return artifactDirs[0]
}

func requireNoReleaseRunnerWorkDir(t *testing.T, repoRoot string) {
	t.Helper()
	requireNoReleaseRunnerWorkDirAtRoot(t, filepath.Join(repoRoot, ".local", "e2e"))
}

func requireNoReleaseRunnerWorkDirAtRoot(t *testing.T, artifactRoot string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(artifactRoot, "release-*-work"))
	require.NoError(t, err)
	require.Empty(t, matches)
}

func requireExitCode(t *testing.T, err error, want int, output []byte) {
	t.Helper()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("command error = %v, want exit code %d\n%s", err, want, output)
	}
	require.Equalf(t, want, exitError.ExitCode(), "output:\n%s", output)
}

func requireProcessFromFileGone(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	require.NoError(t, err)
	deadline := time.Now().Add(5 * time.Second)
	for {
		process, findErr := os.FindProcess(pid)
		if findErr != nil || process.Signal(syscall.Signal(0)) != nil {
			return
		}
		if time.Now().After(deadline) {
			_ = process.Signal(syscall.SIGKILL)
			t.Fatalf("release runner left descendant PID %d alive", pid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func forceKillPIDFromFile(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return
	}
	process, err := os.FindProcess(pid)
	if err == nil {
		_ = process.Signal(syscall.SIGKILL)
	}
}

func TestReleaseDockerfileBasesAreDigestPinned(t *testing.T) {
	for _, dockerfile := range []string{"../Dockerfile", "docker/Dockerfile", "docker/Dockerfile.release"} {
		contents, err := os.ReadFile(dockerfile)
		require.NoError(t, err)
		images, err := harness.ParseReleasePinnedBaseImages(contents)
		require.NoError(t, err)
		require.Len(t, images, 2)
		for _, image := range images {
			require.Regexp(t, `^sha256:[0-9a-f]{64}$`, image.Digest)
		}
	}
}

func TestEveryCurrentReleaseArtifactUsesCanonicalStaticBuildContract(t *testing.T) {
	const buildContract = "panacea-linux-static-v1"

	makefile, err := os.ReadFile("../Makefile")
	require.NoError(t, err)
	require.Contains(t, string(makefile), "release-build:")
	require.Contains(t, string(makefile), "RELEASE_BUILD_CONTRACT := "+buildContract)
	require.Contains(t, string(makefile), "override RELEASE_BUILD_MOD := vendor")
	require.Contains(t, string(makefile), "go build -mod=$(RELEASE_BUILD_MOD)")

	builder, err := os.ReadFile("../scripts/release/build-validator.sh")
	require.NoError(t, err)
	require.Contains(t, string(builder), "release-build")
	require.Contains(t, string(builder), "RELEASE_GOARCH=")
	require.Contains(t, string(builder), "RELEASE_OUTPUT=")
	require.Contains(t, string(builder), "build_contract="+buildContract)
	require.Contains(t, string(builder), "dependency_mode=vendor")
	require.Contains(t, string(builder), `"$validator_go_binary" mod vendor`)
	require.NotContains(t, string(builder), "\n\tinstall\n")

	for _, dockerfile := range []string{"../Dockerfile", "docker/Dockerfile", "docker/Dockerfile.release"} {
		contents, err := os.ReadFile(dockerfile)
		require.NoError(t, err)
		dockerBuild := string(contents)
		require.Contains(t, dockerBuild, "release-build")
		require.Contains(t, dockerBuild, "RELEASE_GOARCH=")
		require.Contains(t, dockerBuild, "RELEASE_OUTPUT=")
		if dockerfile != "docker/Dockerfile" {
			require.Contains(t, dockerBuild, `org.medibloc.panacea.build-contract="`+buildContract+`"`)
		}
	}

	workflow, err := os.ReadFile("../.github/workflows/docker-publish.yml")
	require.NoError(t, err)
	publish := string(workflow)
	require.Contains(t, publish, "platforms: linux/amd64,linux/arm64")
	require.Contains(t, publish, "panacea_vendor=")
	require.Contains(t, publish, "PANACEA_CMT_VERSION=")
	require.NotContains(t, publish, "actions/setup-go@v")
}

func TestCanonicalStaticBuildContractCannotBeDowngradedByMakeArguments(t *testing.T) {
	command := exec.Command(
		"make", "-n", "-C", "..", "release-build",
		"LEDGER_ENABLED=false",
		"VERSION=2.3.0",
		"COMMIT=0123456789abcdef0123456789abcdef01234567",
		"RELEASE_GOARCH=amd64",
		"RELEASE_OUTPUT=/tmp/panacea-release-contract-test",
		"RELEASE_BUILD_MOD=readonly",
		"RELEASE_BUILD_FLAGS=unsafe",
	)
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "make dry-run output:\n%s", output)
	contract := string(output)
	require.Contains(t, contract, `go build -mod=vendor -tags "netgo"`)
	require.NotContains(t, contract, "go build -mod=readonly")
	require.NotContains(t, contract, " unsafe ")
}
