package e2e_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
  version:) printf '%s\n' 'go version go1.26.5 contract/arch' ;;
  env:GOVERSION) printf '%s\n' 'go1.26.5' ;;
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
printf '%s\n' 'compile version go1.26.5'
`)

	command := exec.Command(runner, "unit")
	command.Dir = ".."
	command.Env = append(standaloneRunnerTestEnv(),
		"E2E_ROOT="+testRoot,
		"E2E_GOCACHE="+filepath.Join(testRoot, "go-build"),
		"E2E_GOMODCACHE="+filepath.Join(testRoot, "go-mod"),
		"E2E_GO_VERSION=1.26.5",
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
			environment: []string{"E2E_GO_VERSION=1.26.5", "E2E_GOTOOLCHAIN=auto"},
			diagnostic:  "E2E_GOTOOLCHAIN must be local",
		},
		{
			name:        "different expected Go version",
			environment: []string{"E2E_GO_VERSION=1.25.7", "E2E_GOTOOLCHAIN=local"},
			diagnostic:  "E2E_GO_VERSION must be 1.26.5",
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

func TestCosmovisorRehearsalIsExplicitAndManualOnly(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	runner := string(contents)
	body := shellFunctionBody(t, runner, "cosmovisor_body")

	require.Contains(t, body, `${PANACEA_REHEARSAL_OLD_TAG:-}`)
	require.Contains(t, body, `${PANACEA_REHEARSAL_UPGRADE_NAME:-}`)
	require.Contains(t, body, `if [ "$PANACEA_REHEARSAL_OLD_TAG" = "$PANACEA_REHEARSAL_UPGRADE_NAME" ]`)
	require.Contains(t, body, `GOWORK=off`)
	require.Contains(t, body, `PANACEA_REHEARSAL_ROOT="$rehearsal_root"`)
	require.Contains(t, body, `"$repo_root/scripts/upgrade-local/run.sh" run`)
	functions := shellFunctions(t, runner)
	allReachable := reachableShellFunctionsFrom(functions["all_body"], functions)
	require.False(t, allReachable["cosmovisor_body"])

	for _, testCase := range []struct {
		name        string
		environment []string
		diagnostic  string
	}{
		{
			name:       "missing old tag",
			diagnostic: "requires PANACEA_REHEARSAL_OLD_TAG",
		},
		{
			name:        "missing upgrade name",
			environment: []string{"PANACEA_REHEARSAL_OLD_TAG=v2.2.1"},
			diagnostic:  "requires PANACEA_REHEARSAL_UPGRADE_NAME",
		},
		{
			name: "equal versions",
			environment: []string{
				"PANACEA_REHEARSAL_OLD_TAG=v2.3.0",
				"PANACEA_REHEARSAL_UPGRADE_NAME=v2.3.0",
			},
			diagnostic: "old tag and upgrade name must differ",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			command := exec.Command("../scripts/e2e/run.sh", "cosmovisor")
			command.Env = append(standaloneRunnerTestEnv(), testCase.environment...)
			output, commandErr := command.CombinedOutput()
			requireExitCode(t, commandErr, 2, output)
			require.Contains(t, string(output), testCase.diagnostic)
		})
	}

	ciContents, err := os.ReadFile("../.github/workflows/ci.yml")
	require.NoError(t, err)
	ci := string(ciContents)
	require.NotContains(t, ci, "cosmovisor-rehearsal")
	require.NotContains(t, ci, "workflow_dispatch:")
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

func TestStandaloneRestartAndConsensusSuitesOnlyRetryHostPortRaces(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	runner := string(contents)

	restartBody := shellFunctionBody(t, runner, "restart_body")
	require.Equal(t, 2, strings.Count(restartBody, "run_current_test_with_host_port_retry"),
		"restart recovery scenarios must run in separate test-binary processes")
	require.Contains(t, restartBody, "^TestRestartRecoveryNodeBoundaries$")
	require.Contains(t, restartBody, "^TestPortableApplicationSnapshotRestoreAndFreshFullNodeSync$")
	require.NotContains(t, restartBody,
		"TestRestartRecoveryNodeBoundaries|TestPortableApplicationSnapshotRestoreAndFreshFullNodeSync")

	consensusBody := shellFunctionBody(t, runner, "consensus_body")
	require.Equal(t, 1, strings.Count(consensusBody, "run_current_test_with_host_port_retry"),
		"the validator stop/start scenario must retry a Docker host-port allocation race once")
	require.Contains(t, consensusBody, "^TestFourValidatorQuorumFaultAndRecovery$")

	retryBody := shellFunctionBody(t, runner, "run_current_test_with_host_port_retry")
	require.Contains(t, retryBody, "failed to bind host port .*address already in use")
	require.Contains(t, retryBody, "run_current_test")
	require.Contains(t, retryBody, `"$attempt" -ge 2`,
		"the infrastructure-only retry must be bounded to one retry")
}

func TestStandaloneHostPortRetryBehavior(t *testing.T) {
	runner, err := filepath.Abs("../scripts/e2e/run.sh")
	require.NoError(t, err)

	for _, testCase := range []struct {
		name            string
		commandName     string
		failureMode     string
		wantExit        int
		wantInvocations int
		wantRetry       bool
		firstPattern    string
		lastPattern     string
	}{
		{
			name: "restart exact Docker host-port race", commandName: "restart",
			failureMode: "restart-host-port", wantInvocations: 3, wantRetry: true,
			firstPattern: "^TestRestartRecoveryNodeBoundaries$",
			lastPattern:  "^TestPortableApplicationSnapshotRestoreAndFreshFullNodeSync$",
		},
		{
			name: "restart ordinary test failure", commandName: "restart",
			failureMode: "restart-ordinary", wantExit: 23, wantInvocations: 1,
			firstPattern: "^TestRestartRecoveryNodeBoundaries$",
			lastPattern:  "^TestRestartRecoveryNodeBoundaries$",
		},
		{
			name: "consensus exact Docker host-port race", commandName: "consensus",
			failureMode: "consensus-host-port", wantInvocations: 2, wantRetry: true,
			firstPattern: "^TestFourValidatorQuorumFaultAndRecovery$",
			lastPattern:  "^TestFourValidatorQuorumFaultAndRecovery$",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testRoot, tempErr := os.MkdirTemp("/tmp", "panacea-e2e-restart-retry-")
			require.NoError(t, tempErr)
			t.Cleanup(func() { require.NoError(t, os.RemoveAll(testRoot)) })

			goRoot := filepath.Join(testRoot, "fake-goroot")
			toolDir := filepath.Join(goRoot, "pkg", "tool", "contract_arch")
			require.NoError(t, os.MkdirAll(filepath.Join(goRoot, "bin"), 0o700))
			require.NoError(t, os.MkdirAll(toolDir, 0o700))
			fakeGo := filepath.Join(goRoot, "bin", "go")
			writeExecutable(t, fakeGo, `#!/bin/sh
case "${1:-}:${2:-}" in
  version:) printf '%s\n' 'go version go1.26.5 contract/arch' ;;
  env:GOVERSION) printf '%s\n' 'go1.26.5' ;;
  env:GOROOT) printf '%s\n' "$RUNNER_TEST_GOROOT" ;;
  env:GOTOOLDIR) printf '%s\n' "$RUNNER_TEST_GOTOOLDIR" ;;
  env:GOBIN) printf '%s\n' "$RUNNER_TEST_GOROOT/bin" ;;
  mod:vendor) mkdir -p "$4" ;;
  test:*) exit 0 ;;
  *) printf 'unexpected fake go invocation: %s\n' "$*" >&2; exit 97 ;;
esac
`)
			writeExecutable(t, filepath.Join(toolDir, "compile"), `#!/bin/sh
printf '%s\n' 'compile version go1.26.5'
`)
			fakeDocker := filepath.Join(testRoot, "docker")
			writeExecutable(t, fakeDocker, `#!/bin/sh
case "${1:-}" in
  build) exit 0 ;;
  context) printf '%s\n' 'unix:///tmp/panacea-e2e-fake-docker.sock' ;;
  *) printf 'unexpected fake docker invocation: %s\n' "$*" >&2; exit 98 ;;
esac
`)

			invocationLog := filepath.Join(testRoot, "test-binary-invocations.txt")
			retryMarker := filepath.Join(testRoot, "host-port-retried")
			fakeTestBinary := filepath.Join(testRoot, "panacea-e2e.test")
			writeExecutable(t, fakeTestBinary, `#!/bin/sh
printf '%s\n' "$*" >>"$RUNNER_TEST_INVOCATION_LOG"
case "$*" in
  *TestRestartRecoveryNodeBoundaries*)
    if [ ! -e "$RUNNER_TEST_RETRY_MARKER" ]; then
      : >"$RUNNER_TEST_RETRY_MARKER"
      if [ "$RUNNER_TEST_FAILURE_MODE" = restart-host-port ]; then
        printf '%s\n' 'failed to bind host port 0.0.0.0:55230/tcp: address already in use' >&2
        exit 1
      fi
      printf '%s\n' 'application state mismatch' >&2
      exit 23
    fi
    ;;
  *TestPortableApplicationSnapshotRestoreAndFreshFullNodeSync*) ;;
  *TestFourValidatorQuorumFaultAndRecovery*)
    if [ ! -e "$RUNNER_TEST_RETRY_MARKER" ]; then
      : >"$RUNNER_TEST_RETRY_MARKER"
      printf '%s\n' 'failed to bind host port 0.0.0.0:54274/tcp: address already in use' >&2
      exit 1
    fi
    ;;
  *) printf 'unexpected test pattern: %s\n' "$*" >&2; exit 96 ;;
esac
`)

			command := exec.Command(runner, testCase.commandName)
			command.Dir = ".."
			command.Env = append(standaloneRunnerTestEnv(),
				"E2E_ROOT="+testRoot,
				"E2E_GOCACHE="+filepath.Join(testRoot, "go-build"),
				"E2E_GOMODCACHE="+filepath.Join(testRoot, "go-mod"),
				"E2E_GO_VERSION=1.26.5",
				"E2E_GO_BINARY="+fakeGo,
				"E2E_GOTOOLCHAIN=local",
				"E2E_DOCKER_HOST=unix:///tmp/panacea-e2e-fake-docker.sock",
				"DOCKER="+fakeDocker,
				"RUNNER_TEST_GOROOT="+goRoot,
				"RUNNER_TEST_GOTOOLDIR="+toolDir,
				"RUNNER_TEST_INVOCATION_LOG="+invocationLog,
				"RUNNER_TEST_RETRY_MARKER="+retryMarker,
				"RUNNER_TEST_FAILURE_MODE="+testCase.failureMode,
			)
			output, commandErr := command.CombinedOutput()
			if testCase.wantExit == 0 {
				require.NoErrorf(t, commandErr, "runner output:\n%s", output)
			} else {
				requireExitCode(t, commandErr, testCase.wantExit, output)
			}

			invocations, readErr := os.ReadFile(invocationLog)
			require.NoError(t, readErr)
			records := strings.Split(strings.TrimSpace(string(invocations)), "\n")
			require.Len(t, records, testCase.wantInvocations)
			require.Contains(t, records[0], testCase.firstPattern)
			require.Contains(t, records[len(records)-1], testCase.lastPattern)
			if testCase.wantRetry {
				require.Contains(t, string(output), "retrying live E2E once")
				require.Contains(t, records[1], testCase.firstPattern)
			} else {
				require.NotContains(t, string(output), "retrying live E2E once")
			}
		})
	}
}

func TestStandaloneRunnerReachesEveryOptInLiveTestExactlyOnce(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	runner := string(contents)
	functions := shellFunctions(t, runner)
	reachable := reachableShellFunctions(runner, functions)

	for _, liveTest := range optInLiveTests(t) {
		t.Run(liveTest.name, func(t *testing.T) {
			var owners []string
			for functionName, body := range functions {
				if strings.Contains(body, liveTest.name) {
					owners = append(owners, functionName)
				}
			}
			require.Lenf(t, owners, 1,
				"live test must occur in exactly one canonical runner function; got %v", owners)
			owner := owners[0]
			require.Truef(t, reachable[owner],
				"runner function %s is not reachable from a public command", owner)
			require.Containsf(t, functions[owner], liveTest.flag,
				"runner function %s must enable the test's primary opt-in flag", owner)
		})
	}
}

func TestStandaloneAllRunsEveryFunctionalLiveTestAndFeedsTheReleaseGate(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	runner := string(contents)
	functions := shellFunctions(t, runner)
	allReachable := reachableShellFunctionsFrom(functions["all_body"], functions)

	require.Contains(t, functions["all_body"], "unit_body")
	require.False(t, allReachable["release_builds_body"],
		"the functional aggregate must not silently perform release builds")
	for _, liveTest := range optInLiveTests(t) {
		var owners []string
		for functionName, body := range functions {
			if strings.Contains(body, liveTest.name) {
				owners = append(owners, functionName)
			}
		}
		require.Lenf(t, owners, 1, "%s must have one canonical runner owner", liveTest.name)
		require.Truef(t, allReachable[owners[0]],
			"all does not reach live test %s through %s", liveTest.name, owners[0])
	}

	releaseBody := functions["release_hardening_inner_body"]
	require.Equal(t, strings.Join([]string{
		"check_clean",
		"all_body",
		"release_builds_body",
		"check_clean",
		"coverage_merge",
	}, "\n\t"), strings.TrimSpace(releaseBody),
		"release-hardening must clean-check, run all functional suites, run release builds, recheck source, then merge coverage")

	releaseBuildsBody := functions["release_builds_body"]
	require.Contains(t, releaseBuildsBody,
		`if [ "$current_image_built" -eq 1 ] && [ "$v221_image_built" -eq 1 ]; then`)
	require.Contains(t, releaseBuildsBody,
		`E2E_FUNCTIONAL_IMAGES_PREBUILT="$functional_images_prebuilt"`)
}

type optInLiveTest struct {
	name string
	flag string
}

func optInLiveTests(t *testing.T) []optInLiveTest {
	t.Helper()
	paths, err := filepath.Glob("*_test.go")
	require.NoError(t, err)

	files := token.NewFileSet()
	functions := make(map[string]*ast.FuncDecl)
	var testNames []string
	for _, path := range paths {
		parsed, parseErr := parser.ParseFile(files, path, nil, 0)
		require.NoErrorf(t, parseErr, "parse %s", path)
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || function.Body == nil {
				continue
			}
			functions[function.Name.Name] = function
			if strings.HasPrefix(function.Name.Name, "Test") {
				testNames = append(testNames, function.Name.Name)
			}
		}
	}

	var tests []optInLiveTest
	for _, testName := range testNames {
		if primaryFlag := optInFlagReachableFrom(functions, testName, nil); primaryFlag != "" {
			tests = append(tests, optInLiveTest{name: testName, flag: primaryFlag})
		}
	}
	require.NotEmpty(t, tests, "expected opt-in live tests in the E2E package")
	return tests
}

func optInFlagReachableFrom(functions map[string]*ast.FuncDecl, functionName string, visiting map[string]bool) string {
	function := functions[functionName]
	if function == nil {
		return ""
	}
	if visiting == nil {
		visiting = make(map[string]bool)
	}
	if visiting[functionName] {
		return ""
	}
	visiting[functionName] = true
	defer delete(visiting, functionName)

	var primaryFlag string
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if primaryFlag != "" {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Getenv" && len(call.Args) == 1 {
			packageName, packageOK := selector.X.(*ast.Ident)
			literal, literalOK := call.Args[0].(*ast.BasicLit)
			if packageOK && packageName.Name == "os" && literalOK && literal.Kind == token.STRING {
				flag, unquoteErr := strconv.Unquote(literal.Value)
				if unquoteErr == nil && strings.HasPrefix(flag, "PANACEA_E2E") {
					primaryFlag = flag
					return false
				}
			}
		}
		if called, ok := call.Fun.(*ast.Ident); ok {
			primaryFlag = optInFlagReachableFrom(functions, called.Name, visiting)
			if primaryFlag != "" {
				return false
			}
		}
		return true
	})
	return primaryFlag
}

func shellFunctions(t *testing.T, script string) map[string]string {
	t.Helper()
	declarations := regexp.MustCompile(`(?m)^([A-Za-z_][A-Za-z0-9_]*)\(\) \{\n`).FindAllStringSubmatch(script, -1)
	functions := make(map[string]string, len(declarations))
	for _, declaration := range declarations {
		functions[declaration[1]] = shellFunctionBody(t, script, declaration[1])
	}
	return functions
}

func reachableShellFunctions(script string, functions map[string]string) map[string]bool {
	dispatchStart := strings.LastIndex(script, `case "$command_name" in`)
	if dispatchStart < 0 {
		return nil
	}
	return reachableShellFunctionsFrom(script[dispatchStart:], functions)
}

func reachableShellFunctionsFrom(rootBody string, functions map[string]string) map[string]bool {
	reachable := make(map[string]bool)
	markCalls := func(body string) bool {
		changed := false
		for functionName := range functions {
			called := regexp.MustCompile(`(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(functionName) + `([^A-Za-z0-9_]|$)`).MatchString(body)
			if called && !reachable[functionName] {
				reachable[functionName] = true
				changed = true
			}
		}
		return changed
	}
	markCalls(rootBody)
	for {
		changed := false
		for functionName := range reachable {
			if markCalls(functions[functionName]) {
				changed = true
			}
		}
		if !changed {
			return reachable
		}
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
		"PANACEA_REHEARSAL_", "RUNNER_TEST_",
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
