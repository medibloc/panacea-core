package e2e_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootGoPackagePatternExcludesAllStandaloneE2EPackages(t *testing.T) {
	command := exec.Command("go", "list", "./...")
	command.Dir = ".."
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "root go list ./... failed:\n%s", output)

	for _, packagePath := range strings.Fields(string(output)) {
		require.Falsef(t,
			strings.HasSuffix(packagePath, "/e2e") || strings.Contains(packagePath, "/e2e/"),
			"root go list ./... must exclude standalone E2E package %s",
			packagePath,
		)
	}
}

func TestOrdinaryDockerBuildContextExcludesStandaloneE2ESource(t *testing.T) {
	contents, err := os.ReadFile("../.dockerignore")
	require.NoError(t, err)

	ignored := make(map[string]bool)
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			ignored[strings.TrimPrefix(line, "/")] = true
		}
	}
	require.True(t, ignored["e2e"], "ordinary Docker builds must exclude the E2E module")
	require.True(t, ignored["scripts/e2e"], "ordinary Docker builds must exclude E2E runner sources")
}
