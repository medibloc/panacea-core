package cmd_test

import (
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
	"time"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	panaceacmd "github.com/medibloc/panacea-core/v2/cmd/panacead/cmd"
)

func TestNewRootCmdCommandTree(t *testing.T) {
	var root *cobra.Command
	require.NotPanics(t, func() {
		root, _ = panaceacmd.NewRootCmd()
	})
	require.NotNil(t, root)

	t.Run("query has one IBC transfer command", func(t *testing.T) {
		query := requireDirectChild(t, root, "query")
		requireDirectChild(t, query, "ibc-transfer")
	})

	t.Run("tx has one IBC transfer command", func(t *testing.T) {
		tx := requireDirectChild(t, root, "tx")
		requireDirectChild(t, tx, "ibc-transfer")
	})

	t.Run("keeps the legacy parameter change proposal command", func(t *testing.T) {
		requireCommandPath(t, root, "tx", "gov", "submit-legacy-proposal", "param-change")
	})

	t.Run("keeps Panacea module commands", func(t *testing.T) {
		expectedPaths := [][]string{
			{"query", "aol", "get-topic"},
			{"tx", "aol", "create-topic"},
			{"query", "did", "get-did"},
			{"tx", "did", "create-did"},
			{"query", "pnft", "get-pnft"},
			{"tx", "pnft", "mint-pnft"},
		}

		for _, path := range expectedPaths {
			requireCommandPath(t, root, path...)
		}
	})

	t.Run("keeps SDK module commands", func(t *testing.T) {
		requireCommandPath(t, root, "query", "auth", "account")
		requireCommandPath(t, root, "tx", "staking", "delegate")
	})

	t.Run("has unique command names", func(t *testing.T) {
		requireUniqueChildNames(t, root)
	})

	t.Run("init writes application genesis", func(t *testing.T) {
		home := t.TempDir()
		root.SetArgs([]string{
			"init",
			"test-node",
			"--home", home,
			"--chain-id", "test-chain",
			"--initial-height", "7",
		})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)

		require.NoError(t, svrcmd.Execute(root, "", home))

		appGenesis, err := genutiltypes.AppGenesisFromFile(filepath.Join(home, "config", "genesis.json"))
		require.NoError(t, err)
		require.Equal(t, "test-chain", appGenesis.ChainID)
		require.Equal(t, int64(7), appGenesis.InitialHeight)
		require.False(t, appGenesis.GenesisTime.IsZero())
		require.NotNil(t, appGenesis.Consensus)
		require.NotNil(t, appGenesis.Consensus.Params)
		require.Empty(t, appGenesis.Consensus.Validators)
		require.Equal(t, 21*24*time.Hour, appGenesis.Consensus.Params.Evidence.MaxAgeDuration)
		require.Equal(t, int64(21*24*time.Hour/(6*time.Second)), appGenesis.Consensus.Params.Evidence.MaxAgeNumBlocks)

		var appState map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(appGenesis.AppState, &appState))
		require.NotEmpty(t, appState)
	})
}

func requireUniqueChildNames(t *testing.T, parent *cobra.Command) {
	t.Helper()

	seen := make(map[string]struct{}, len(parent.Commands()))
	for _, child := range parent.Commands() {
		_, duplicate := seen[child.Name()]
		require.False(t, duplicate, "duplicate command %q under %q", child.Name(), parent.CommandPath())

		seen[child.Name()] = struct{}{}
		requireUniqueChildNames(t, child)
	}
}

func requireCommandPath(t *testing.T, root *cobra.Command, path ...string) *cobra.Command {
	t.Helper()

	current := root
	for _, name := range path {
		current = requireDirectChild(t, current, name)
	}
	return current
}

func requireDirectChild(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()

	var matches []*cobra.Command
	for _, child := range parent.Commands() {
		if child.Name() == name {
			matches = append(matches, child)
		}
	}

	require.Len(t, matches, 1, "expected exactly one %q command under %q", name, parent.CommandPath())
	return matches[0]
}
