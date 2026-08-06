package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStageGitSourceMaterializesOnlyTheCommittedTree(t *testing.T) {
	scriptPath, err := filepath.Abs("../scripts/e2e/stage-git-source.sh")
	require.NoError(t, err)

	repo := t.TempDir()
	runGit(t, repo, "init", "--quiet")
	runGit(t, repo, "config", "user.name", "Panacea E2E")
	runGit(t, repo, "config", "user.email", "panacea-e2e@example.invalid")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "cmd", "panacead"), 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, ".gitignore"),
		[]byte("ignored-*.go\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, "go.mod"),
		[]byte("module example.invalid/panacea\n\ngo 1.26.5\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, "cmd", "panacead", "main.go"),
		[]byte("package main\n\nfunc main() {}\n"),
		0o600,
	))
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "--quiet", "-m", "committed source")

	ignoredPath := filepath.Join(repo, "cmd", "panacead", "ignored-backdoor.go")
	untrackedPath := filepath.Join(repo, "cmd", "panacead", "untracked.go")
	require.NoError(t, os.WriteFile(ignoredPath, []byte("package main\n"), 0o600))
	require.NoError(t, os.WriteFile(untrackedPath, []byte("package main\n"), 0o600))

	destination := t.TempDir()
	command := exec.Command("sh", scriptPath, "HEAD", destination)
	command.Dir = repo
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	require.Equal(t, runGit(t, repo, "rev-parse", "HEAD"), strings.TrimSpace(string(output)))

	expected := strings.Fields(runGit(t, repo, "ls-tree", "-r", "--name-only", "HEAD"))
	actual := make([]string, 0, len(expected))
	require.NoError(t, filepath.WalkDir(destination, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, relErr := filepath.Rel(destination, path)
		if relErr != nil {
			return relErr
		}
		actual = append(actual, filepath.ToSlash(relative))
		return nil
	}))
	sort.Strings(expected)
	sort.Strings(actual)
	require.Equal(t, expected, actual)
	require.NoFileExists(t, filepath.Join(destination, "cmd", "panacead", filepath.Base(ignoredPath)))
	require.NoFileExists(t, filepath.Join(destination, "cmd", "panacead", filepath.Base(untrackedPath)))
}

func TestCurrentFunctionalImageUsesCommitSourceOnlyWithReleaseOverride(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	runner := string(contents)

	for _, contract := range []string{
		`E2E_CURRENT_SOURCE_COMMIT=${E2E_CURRENT_SOURCE_COMMIT:-}`,
		`current_source_dir=$repo_root`,
		`if [ -n "$E2E_CURRENT_SOURCE_COMMIT" ]`,
		`stage-git-source.sh" "$E2E_CURRENT_SOURCE_COMMIT" "$current_source_dir"`,
		`cd "$current_source_dir"`,
		`mod vendor -o "$stage_dir/vendor"`,
		`--build-context panacea_e2e_tools="$current_tools_dir"`,
		`--file "$current_dockerfile"`,
		`--build-arg PANACEA_COMMIT="$build_commit"`,
		`"$current_source_dir"`,
	} {
		require.Contains(t, runner, contract)
	}
}

func TestReleaseRunnerUsesOnePinnedCurrentSourceForEveryBuildBoundary(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/release-hardening.sh")
	require.NoError(t, err)
	script := string(contents)

	for _, contract := range []string{
		`stage=prepare-current-source`,
		`current_source="$work_dir/current-source"`,
		`stage-git-source.sh" "$source_commit" "$current_source"`,
		`E2E_CURRENT_SOURCE_COMMIT="$source_commit"`,
		`cd "$current_source"`,
		`cd "$current_source/e2e"`,
		`dockerfile="$current_source/e2e/docker/Dockerfile.release"`,
		`--build-context "panacea_e2e_tools=$current_source/scripts/e2e"`,
		`build_and_verify_image "$platform" "$suffix" current "$current_source" "$current_vendor"`,
		`warm_offline_buildkit_image "$platform" "$suffix" current "$current_source" "$current_vendor"`,
	} {
		require.Contains(t, script, contract)
	}
	require.NotContains(t, script, `current "$repo_root" "$current_vendor"`)
	require.Contains(t, script, `git archive --format=tar --output="$work_dir/v2.2.1-source.tar" "$E2E_V221_COMMIT"`)
	require.Contains(t, script, `build_and_verify_image "$platform" "$suffix" v2.2.1 "$old_source" "$old_vendor"`)
}

func TestOldFunctionalImageKeepsTaggedNodeSourceAndStagesCurrentBuildToolsForRelease(t *testing.T) {
	contents, err := os.ReadFile("../scripts/e2e/run.sh")
	require.NoError(t, err)
	runner := string(contents)

	for _, contract := range []string{
		`git archive --format=tar --output="$stage_dir/source.tar" "$E2E_V221_COMMIT"`,
		`v221_tools_dir=$script_dir`,
		`if [ -n "$E2E_CURRENT_SOURCE_COMMIT" ]`,
		`tools_source_dir="$stage_dir/current-source"`,
		`stage-git-source.sh" "$E2E_CURRENT_SOURCE_COMMIT" "$tools_source_dir"`,
		`v221_tools_dir="$tools_source_dir/scripts/e2e"`,
		`v221_dockerfile="$tools_source_dir/e2e/docker/Dockerfile"`,
		`--build-context panacea_e2e_tools="$v221_tools_dir"`,
		`--build-arg PANACEA_COMMIT="$E2E_V221_COMMIT"`,
		`"$v221_source_dir"`,
	} {
		require.Contains(t, runner, contract)
	}
}

func runGit(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "git %s failed:\n%s", strings.Join(arguments, " "), output)
	return strings.TrimSpace(string(output))
}
