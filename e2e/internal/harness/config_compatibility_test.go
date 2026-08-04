package harness

import (
	"fmt"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

const legacyAppConfigFixture = `app-db-backend = ""
minimum-gas-prices = "5umed"
query-gas-limit = 10000000

[api]
  address = "tcp://0.0.0.0:1317"
  enable = true
  max-open-connections = 1000
  rpc-max-body-bytes = 1000000
  rpc-read-timeout = 10
  rpc-write-timeout = 0

[grpc]
  address = "0.0.0.0:9090"
  enable = true
  max-recv-msg-size = "10485760"
  max-send-msg-size = "2147483647"

[grpc-web]
  address = "localhost:9091"
  enable = true
  enable-unsafe-cors = false
`

const legacyClientConfigFixture = `chain-id = "panacea-config-test"
keyring-backend = "os"
output = "text"
node = "tcp://localhost:26657"
broadcast-mode = "sync"
`

const legacyCometConfigFixture = `db_backend = "goleveldb"

[mempool]
  max_txs_bytes = 1073741824
  recheck = true
  size = 5000
  type = "flood"
  version = "v0"

[p2p]
  dial_timeout = "3s"
  laddr = "tcp://0.0.0.0:26656"
  persistent_peers = "abc@validator:26656"

[rpc]
  laddr = "tcp://0.0.0.0:26657"
  timeout_broadcast_tx_commit = "10s"

[tx_index]
  indexer = "kv"
`

func TestPrepareV047NodeHomeConfigAppliesAuditableOverrides(t *testing.T) {
	prepared, err := PrepareV047NodeHomeConfig(
		[]byte(legacyAppConfigFixture),
		[]byte(legacyClientConfigFixture),
		[]byte(legacyCometConfigFixture),
		"panacea-config-test",
	)
	require.NoError(t, err)
	require.NoError(t, prepared.Validate())

	require.Equal(t, int64(configCompatQueryGasLimit), tomlInt64(t, prepared.App, "query-gas-limit"))
	require.Equal(t, int64(configCompatAPIReadTimeout), tomlInt64(t, prepared.App, "api.rpc-read-timeout"))
	require.Equal(t, int64(configCompatAPIWriteTimeout), tomlInt64(t, prepared.App, "api.rpc-write-timeout"))
	require.Equal(t, int64(configCompatAPIMaxBodyBytes), tomlInt64(t, prepared.App, "api.rpc-max-body-bytes"))
	require.Equal(t, configCompatGRPCMaxRecvMsgSize, tomlString(t, prepared.App, "grpc.max-recv-msg-size"))
	require.Equal(t, configCompatGRPCMaxSendMsgSize, tomlString(t, prepared.App, "grpc.max-send-msg-size"))
	require.Equal(t, "json", tomlString(t, prepared.Client, "output"))
	require.Equal(t, int64(configCompatMempoolSize), tomlInt64(t, prepared.Comet, "mempool.size"))
	require.Equal(t, int64(configCompatMempoolMaxTxsBytes), tomlInt64(t, prepared.Comet, "mempool.max_txs_bytes"))
	require.Equal(t, configCompatP2PDialTimeout, tomlString(t, prepared.Comet, "p2p.dial_timeout"))
	require.Equal(t, configCompatRPCTxCommitTimeout, tomlString(t, prepared.Comet, "rpc.timeout_broadcast_tx_commit"))
	require.Equal(t, "abc@validator:26656", tomlString(t, prepared.Comet, "p2p.persistent_peers"))
	require.NotEmpty(t, prepared.SHA256["app.toml"])
	require.NotEmpty(t, prepared.SHA256["client.toml"])
	require.NotEmpty(t, prepared.SHA256["config.toml"])
}

func TestValidatePreservedV047NodeHomeRequiresByteIdentity(t *testing.T) {
	prepared := mustPreparedV047NodeHome(t)
	require.NoError(t, ValidatePreservedV047NodeHome(prepared, prepared))

	mutated := prepared
	mutated.Client = append(append([]byte(nil), prepared.Client...), '\n')
	mutated.SHA256 = NodeHomeConfigDigests(mutated.App, mutated.Client, mutated.Comet)
	require.ErrorContains(t, ValidatePreservedV047NodeHome(prepared, mutated), "client.toml changed")
}

func TestValidateMigratedV050NodeHomePreservesOverridesAndRemovesLegacyGRPCWebKeys(t *testing.T) {
	before := mustPreparedV047NodeHome(t)
	after := NewNodeHomeConfigSnapshot(migratedV050App(t, before.App), before.Client, before.Comet)

	require.NoError(t, ValidateMigratedV050NodeHome(before, after))
}

func TestValidateMigratedV050NodeHomeRejectsSilentWeakeningOrNonAppMutation(t *testing.T) {
	before := mustPreparedV047NodeHome(t)
	migratedApp := string(migratedV050App(t, before.App))
	migratedApp = strings.ReplaceAll(migratedApp, "query-gas-limit = 7654321", "query-gas-limit = 0")
	after := NewNodeHomeConfigSnapshot([]byte(migratedApp), before.Client, before.Comet)
	require.ErrorContains(t, ValidateMigratedV050NodeHome(before, after), "query-gas-limit")

	after = NewNodeHomeConfigSnapshot(
		[]byte(strings.ReplaceAll(migratedApp, "query-gas-limit = 0", "query-gas-limit = 7654321")),
		append(append([]byte(nil), before.Client...), '\n'),
		before.Comet,
	)
	require.ErrorContains(t, ValidateMigratedV050NodeHome(before, after), "client.toml changed")
}

func TestValidateMigratedV050NodeHomeRejectsUnexpectedSemanticDiff(t *testing.T) {
	before := mustPreparedV047NodeHome(t)
	migratedTree, err := toml.LoadBytes(migratedV050App(t, before.App))
	require.NoError(t, err)
	migratedTree.SetPath([]string{"api", "unexpected-release-switch"}, true)
	after := NewNodeHomeConfigSnapshot([]byte(migratedTree.String()), before.Client, before.Comet)
	require.ErrorContains(t, ValidateMigratedV050NodeHome(before, after), "unexpected added path")

	migratedTree.DeletePath([]string{"api", "unexpected-release-switch"})
	migratedTree.DeletePath([]string{"minimum-gas-prices"})
	after = NewNodeHomeConfigSnapshot([]byte(migratedTree.String()), before.Client, before.Comet)
	require.ErrorContains(t, ValidateMigratedV050NodeHome(before, after), "unexpected removed path")

	migratedTree.SetPath([]string{"minimum-gas-prices"}, "5umed")
	migratedTree.SetPath([]string{"telemetry", "metrics-sink"}, "statsd")
	after = NewNodeHomeConfigSnapshot([]byte(migratedTree.String()), before.Client, before.Comet)
	require.ErrorContains(t, ValidateMigratedV050NodeHome(before, after), "v0.50 default")
}

func TestValidateMigratedV050NodeHomeRejectsUnsupportedDatabaseBackend(t *testing.T) {
	before := mustPreparedV047NodeHome(t)
	badComet := strings.ReplaceAll(string(before.Comet), `db_backend = "goleveldb"`, `db_backend = "badgerdb"`)
	after := NewNodeHomeConfigSnapshot(migratedV050App(t, before.App), before.Client, []byte(badComet))
	require.ErrorContains(t, ValidateMigratedV050NodeHome(before, after), "unsupported")
	require.ErrorContains(t, after.Validate(), "unsupported")
}

func migratedV050App(t *testing.T, contents []byte) []byte {
	t.Helper()
	tree, err := toml.LoadBytes(contents)
	require.NoError(t, err)
	tree.DeletePath([]string{"grpc-web", "address"})
	tree.DeletePath([]string{"grpc-web", "enable-unsafe-cors"})
	for path := range configCompatV050RemovedPaths {
		tree.DeletePath(strings.Split(path, "."))
	}
	for path, value := range configCompatV050AddedValues {
		tree.SetPath(strings.Split(path, "."), value)
	}
	return []byte(tree.String())
}

func mustPreparedV047NodeHome(t *testing.T) NodeHomeConfigSnapshot {
	t.Helper()
	prepared, err := PrepareV047NodeHomeConfig(
		[]byte(legacyAppConfigFixture),
		[]byte(legacyClientConfigFixture),
		[]byte(legacyCometConfigFixture),
		"panacea-config-test",
	)
	require.NoError(t, err)
	return prepared
}

func tomlTestValue(t *testing.T, contents []byte, path string) any {
	t.Helper()
	tree, err := toml.LoadBytes(contents)
	require.NoError(t, err)
	value := tree.GetPath(strings.Split(path, "."))
	require.NotNil(t, value, "TOML path %s", path)
	return value
}

func tomlInt64(t *testing.T, contents []byte, path string) int64 {
	t.Helper()
	parsed, err := positiveTOMLInteger(tomlTestValue(t, contents, path))
	require.NoError(t, err)
	return parsed
}

func tomlString(t *testing.T, contents []byte, path string) string {
	t.Helper()
	return fmt.Sprint(tomlTestValue(t, contents, path))
}
