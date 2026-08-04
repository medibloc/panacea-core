package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStandaloneUnitCommandUsesPinnedNestedModuleBoundary(t *testing.T) {
	runner, err := filepath.Abs("../scripts/e2e/run.sh")
	require.NoError(t, err)

	testRoot, err := os.MkdirTemp("/tmp", "panacea-e2e-runner-contract-")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(testRoot)) })

	goRoot := filepath.Join(testRoot, "fake-goroot")
	toolDir := filepath.Join(goRoot, "pkg", "tool", "contract_arch")
	require.NoError(t, os.MkdirAll(filepath.Join(goRoot, "bin"), 0o700))
	require.NoError(t, os.MkdirAll(toolDir, 0o700))
	goLog := filepath.Join(testRoot, "go-test-invocation.txt")
	fakeGo := filepath.Join(goRoot, "bin", "go")
	writeExecutable(t, fakeGo, `#!/bin/sh
case "${1:-}:${2:-}" in
  version:) printf '%s\n' 'go version go1.23.12 contract/arch' ;;
  env:GOVERSION) printf '%s\n' 'go1.23.12' ;;
  env:GOROOT) printf '%s\n' "$RUNNER_TEST_GOROOT" ;;
  env:GOTOOLDIR) printf '%s\n' "$RUNNER_TEST_GOTOOLDIR" ;;
  env:GOBIN) printf '%s\n' "$RUNNER_TEST_GOROOT/bin" ;;
  test:*)
    printf 'GOTOOLCHAIN=%s\tGOWORK=%s\tPWD=%s\tARGS=%s\n' \
      "${GOTOOLCHAIN:-}" "${GOWORK:-}" "$PWD" "$*" \
      >>"$RUNNER_TEST_GO_LOG"
    ;;
  *) printf 'unexpected fake go invocation: %s\n' "$*" >&2; exit 97 ;;
esac
`)
	writeExecutable(t, filepath.Join(toolDir, "compile"), `#!/bin/sh
printf '%s\n' 'compile version go1.23.12'
`)

	command := exec.Command(runner, "unit")
	command.Dir = ".."
	command.Env = append(standaloneRunnerTestEnv(),
		"E2E_ROOT="+testRoot,
		"E2E_GOCACHE="+filepath.Join(testRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(testRoot, "go-mod"),
		"E2E_GO_VERSION=1.23.12",
		"E2E_GO_BINARY="+fakeGo,
		"E2E_GOTOOLCHAIN=local",
		"RUNNER_TEST_GOROOT="+goRoot,
		"RUNNER_TEST_GOTOOLDIR="+toolDir,
		"RUNNER_TEST_GO_LOG="+goLog,
	)
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "runner output:\n%s", output)

	invocations, err := os.ReadFile(goLog)
	require.NoError(t, err)
	records := strings.Split(strings.TrimSpace(string(invocations)), "\n")
	require.Len(t, records, 2, "unit must test both nested E2E modules")
	repoRoot, err := filepath.Abs("..")
	require.NoError(t, err)
	repoRoot, err = filepath.EvalSymlinks(repoRoot)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{
		"GOTOOLCHAIN=local\tGOWORK=off\tPWD=" + filepath.Join(repoRoot, "e2e") +
			"\tARGS=test -count=1 ./...",
		"GOTOOLCHAIN=local\tGOWORK=off\tPWD=" + filepath.Join(repoRoot, "scripts", "e2e", "faultproxy") +
			"\tARGS=test -count=1 ./...",
	}, records)
}

func TestStandaloneRunnerRejectsToolchainPolicyOverridesBeforeWork(t *testing.T) {
	runner, err := filepath.Abs("../scripts/e2e/run.sh")
	require.NoError(t, err)

	for _, testCase := range []struct {
		name        string
		environment []string
		diagnostic  string
	}{
		{
			name:        "automatic toolchain selection",
			environment: []string{"E2E_GO_VERSION=1.23.12", "E2E_GOTOOLCHAIN=auto"},
			diagnostic:  "E2E_GOTOOLCHAIN must be local",
		},
		{
			name:        "different expected Go version",
			environment: []string{"E2E_GO_VERSION=1.25.7", "E2E_GOTOOLCHAIN=local"},
			diagnostic:  "E2E_GO_VERSION must be 1.23.12",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			command := exec.Command(runner, "check")
			command.Dir = ".."
			command.Env = append(standaloneRunnerTestEnv(), testCase.environment...)
			output, commandErr := command.CombinedOutput()
			requireExitCode(t, commandErr, 2, output)
			require.Contains(t, string(output), testCase.diagnostic)
		})
	}
}

func TestStandaloneLiveSuitesRunThePrecompiledTestBinaryDirectly(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	runner := string(contents)

	for _, functionName := range []string{"run_current_test", "run_v221_test", "run_upgrade_test"} {
		body := shellFunctionBody(t, runner, functionName)
		require.Containsf(t, body, "build_test_binary", "%s must compile the shared test binary", functionName)
		require.Containsf(t, body, `"$E2E_ROOT/panacea-e2e.test"`, "%s must execute the binary directly", functionName)
		require.Containsf(t, body, "-test.timeout", "%s must preserve the suite timeout", functionName)
		require.NotContainsf(t, body, `"$E2E_GO_BINARY" test -timeout`, "%s must not execute live tests through the go test driver", functionName)
	}
}

func shellFunctionBody(t *testing.T, script, functionName string) string {
	t.Helper()
	startMarker := functionName + "() {\n"
	start := strings.Index(script, startMarker)
	require.NotEqualf(t, -1, start, "missing shell function %s", functionName)
	remainder := script[start+len(startMarker):]
	end := strings.Index(remainder, "\n}\n")
	require.NotEqualf(t, -1, end, "unterminated shell function %s", functionName)
	return remainder[:end]
}

func standaloneRunnerTestEnv() []string {
	prefixes := []string{
		"E2E_ROOT=", "E2E_GOCACHE=", "E2E_GOMODCACHE=", "E2E_GO_VERSION=",
		"E2E_GO_BINARY=", "E2E_GOTOOLCHAIN=", "GOTOOLCHAIN=", "GOWORK=",
		"RUNNER_TEST_",
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
