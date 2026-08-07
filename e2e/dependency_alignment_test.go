package e2e_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/mod/modfile"

	"github.com/stretchr/testify/require"
)

var coreStackModulePaths = []string{
	"cosmossdk.io/api",
	"cosmossdk.io/collections",
	"cosmossdk.io/core",
	"cosmossdk.io/depinject",
	"cosmossdk.io/errors",
	"cosmossdk.io/log",
	"cosmossdk.io/math",
	"cosmossdk.io/store",
	"cosmossdk.io/x/evidence",
	"cosmossdk.io/x/feegrant",
	"cosmossdk.io/x/nft",
	"cosmossdk.io/x/tx",
	"cosmossdk.io/x/upgrade",
	"github.com/cometbft/cometbft",
	"github.com/cometbft/cometbft-db",
	"github.com/cosmos/cosmos-db",
	"github.com/cosmos/cosmos-proto",
	"github.com/cosmos/cosmos-sdk",
	"github.com/cosmos/gogogateway",
	"github.com/cosmos/gogoproto",
	"github.com/cosmos/iavl",
	"github.com/cosmos/ibc-go/modules/capability",
	"github.com/cosmos/ibc-go/v8",
	"github.com/cosmos/ics23/go",
	"github.com/cosmos/ledger-cosmos-go",
	"google.golang.org/genproto/googleapis/api",
	"google.golang.org/genproto/googleapis/rpc",
	"google.golang.org/grpc",
	"google.golang.org/protobuf",
}

func TestE2ECoreStackMatchesRootModule(t *testing.T) {
	t.Parallel()

	_, sourceFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	e2eRoot := filepath.Dir(sourceFile)

	rootVersions := readRequiredModuleVersions(t, filepath.Join(e2eRoot, "..", "go.mod"))
	e2eVersions := readRequiredModuleVersions(t, filepath.Join(e2eRoot, "go.mod"))
	for _, modulePath := range coreStackModulePaths {
		rootVersion, rootExists := rootVersions[modulePath]
		require.Truef(t, rootExists, "root go.mod must require core module %s", modulePath)
		e2eVersion, e2eExists := e2eVersions[modulePath]
		require.Truef(t, e2eExists, "e2e/go.mod must require core module %s", modulePath)
		require.Equalf(t, rootVersion, e2eVersion, "core module %s must match the root go.mod", modulePath)
	}
}

func readRequiredModuleVersions(t *testing.T, path string) map[string]string {
	t.Helper()

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	parsed, err := modfile.Parse(path, contents, nil)
	require.NoError(t, err)

	versions := make(map[string]string, len(parsed.Require))
	for _, requirement := range parsed.Require {
		versions[requirement.Mod.Path] = requirement.Mod.Version
	}
	return versions
}
