package harness

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EPathPreflight(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", ".."))
	script := filepath.Join(repositoryRoot, "scripts", "e2e", "validate-paths.sh")

	run := func(root, goCache, goModCache string) ([]byte, error) {
		command := exec.Command("sh", script)
		command.Env = append(os.Environ(),
			"TMPDIR=/",
			"E2E_ROOT="+root,
			"E2E_GOCACHE="+goCache,
			"E2E_GOMODCACHE="+goModCache,
		)
		return command.CombinedOutput()
	}

	repositoryE2ERoot := filepath.Join(repositoryRoot, ".local", "e2e", "path-preflight-test")
	output, err := run(
		repositoryE2ERoot,
		filepath.Join(repositoryE2ERoot, "go-build"),
		filepath.Join(repositoryE2ERoot, "go-mod"),
	)
	require.NoError(t, err, string(output))

	temporaryRoot := filepath.Join(trustedArtifactTempDir(t), "e2e")
	output, err = run(temporaryRoot, filepath.Join(temporaryRoot, "go-build"), filepath.Join(temporaryRoot, "go-mod"))
	require.NoError(t, err, string(output))

	output, err = run(
		filepath.Join(repositoryRoot, "docs"),
		filepath.Join(repositoryRoot, "docs", "go-build"),
		filepath.Join(repositoryRoot, "docs", "go-mod"),
	)
	require.Error(t, err)
	require.Contains(t, string(output), "must resolve under")

	root := filepath.Join(trustedArtifactTempDir(t), "root")
	output, err = run(root, filepath.Join(trustedArtifactTempDir(t), "go-build"), filepath.Join(root, "go-mod"))
	require.Error(t, err)
	require.Contains(t, string(output), "E2E_GOCACHE must resolve under E2E_ROOT")

	output, err = run(
		filepath.Join(repositoryRoot, ".local", "e2e")+"/../../docs",
		filepath.Join(repositoryRoot, ".local", "e2e", "go-build"),
		filepath.Join(repositoryRoot, ".local", "e2e", "go-mod"),
	)
	require.Error(t, err)
	require.Contains(t, string(output), "must not contain . or .. components")

	symlinkBase := trustedArtifactTempDir(t)
	symlinkRoot := filepath.Join(symlinkBase, "root")
	symlinkTarget := filepath.Join(symlinkBase, "outside")
	require.NoError(t, os.MkdirAll(symlinkRoot, 0o700))
	require.NoError(t, os.MkdirAll(symlinkTarget, 0o700))
	require.NoError(t, os.Symlink(symlinkTarget, filepath.Join(symlinkRoot, "go-build")))
	output, err = run(symlinkRoot, filepath.Join(symlinkRoot, "go-build"), filepath.Join(symlinkRoot, "go-mod"))
	require.Error(t, err)
	require.Contains(t, string(output), "E2E_GOCACHE must resolve under E2E_ROOT")
}
