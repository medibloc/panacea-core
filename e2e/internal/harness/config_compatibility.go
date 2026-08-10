package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml"
)

const (
	configCompatQueryGasLimit      = 7_654_321
	configCompatAPIReadTimeout     = 13
	configCompatAPIWriteTimeout    = 17
	configCompatAPIMaxBodyBytes    = 765_432
	configCompatAPIMaxConnections  = 432
	configCompatGRPCMaxRecvMsgSize = "8388608"
	configCompatGRPCMaxSendMsgSize = "9437184"
	configCompatMempoolSize        = 4_321
	configCompatMempoolMaxTxsBytes = 67_108_864
	configCompatP2PDialTimeout     = "4s"
	configCompatRPCTxCommitTimeout = "13s"
	configCompatRPCListener        = "tcp://0.0.0.0:26657"
	configCompatP2PListener        = "tcp://0.0.0.0:26656"
	configCompatAPIListener        = "tcp://0.0.0.0:1317"
	configCompatApplicationBackend = ""
	configCompatCometBackend       = "goleveldb"
)

var configCompatUnsupportedDBBackends = map[string]struct{}{
	"badgerdb": {},
	"boltdb":   {},
	"cleveldb": {},
}

// Confix v0.1.2 migrates v0.47 -> v0.50 by changing the app.toml keyspace,
// not by replacing values which exist in both schemas. Keep this contract
// explicit so an unexpected key deletion, insertion, or value rewrite cannot
// hide behind a handful of preserved operator overrides.
var configCompatV050AddedValues = map[string]any{
	"telemetry.datadog-hostname":      "",
	"telemetry.metrics-sink":          "mem",
	"telemetry.statsd-addr":           "",
	"streaming.abci.keys":             []any{},
	"streaming.abci.plugin":           "",
	"streaming.abci.stop-node-on-err": true,
}

var configCompatV050RemovedPaths = map[string]struct{}{
	"iavl-lazy-loading":                 {},
	"grpc-web.address":                  {},
	"grpc-web.enable-unsafe-cors":       {},
	"rosetta.enable":                    {},
	"rosetta.address":                   {},
	"rosetta.blockchain":                {},
	"rosetta.network":                   {},
	"rosetta.retries":                   {},
	"rosetta.offline":                   {},
	"rosetta.enable-fee-suggestion":     {},
	"rosetta.gas-to-suggest":            {},
	"rosetta.denom-to-suggest":          {},
	"store.streamers":                   {},
	"streamers.file.keys":               {},
	"streamers.file.write_dir":          {},
	"streamers.file.prefix":             {},
	"streamers.file.output-metadata":    {},
	"streamers.file.stop-node-on-error": {},
	"streamers.file.fsync":              {},
}

// NodeHomeConfigSnapshot is the byte-exact node-local configuration contract.
// It deliberately excludes genesis, keys, address books, data, and every other
// potentially sensitive node-home file.
type NodeHomeConfigSnapshot struct {
	App    []byte            `json:"-"`
	Client []byte            `json:"-"`
	Comet  []byte            `json:"-"`
	SHA256 map[string]string `json:"sha256"`
}

// NewNodeHomeConfigSnapshot owns copies of the three public TOML documents and
// computes stable digests suitable for artifact manifests.
func NewNodeHomeConfigSnapshot(app, client, comet []byte) NodeHomeConfigSnapshot {
	appCopy := append([]byte(nil), app...)
	clientCopy := append([]byte(nil), client...)
	cometCopy := append([]byte(nil), comet...)
	return NodeHomeConfigSnapshot{
		App:    appCopy,
		Client: clientCopy,
		Comet:  cometCopy,
		SHA256: NodeHomeConfigDigests(appCopy, clientCopy, cometCopy),
	}
}

// NodeHomeConfigDigests returns a fixed-key digest map so a missing document
// cannot silently disappear from the release evidence.
func NodeHomeConfigDigests(app, client, comet []byte) map[string]string {
	return map[string]string{
		"app.toml":    configCompatibilitySHA256Hex(app),
		"client.toml": configCompatibilitySHA256Hex(client),
		"config.toml": configCompatibilitySHA256Hex(comet),
	}
}

func configCompatibilitySHA256Hex(contents []byte) string {
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:])
}

// Validate proves that all three files are non-empty TOML, the stored digests
// describe those exact bytes, and the application/database backend fallback is
// safe for Cosmos SDK v0.50.
func (s NodeHomeConfigSnapshot) Validate() error {
	documents := map[string][]byte{
		"app.toml":    s.App,
		"client.toml": s.Client,
		"config.toml": s.Comet,
	}
	for name, contents := range documents {
		if len(bytes.TrimSpace(contents)) == 0 {
			return fmt.Errorf("%s is empty", name)
		}
		if _, err := toml.LoadBytes(contents); err != nil {
			return fmt.Errorf("parse %s: %w", name, err)
		}
		want := configCompatibilitySHA256Hex(contents)
		if got := strings.TrimSpace(s.SHA256[name]); got != want {
			return fmt.Errorf("%s SHA-256 %q, want %q", name, got, want)
		}
	}
	if len(s.SHA256) != len(documents) {
		return fmt.Errorf("node-home digest map contains %d files, want %d", len(s.SHA256), len(documents))
	}
	return validateConfigCompatibilityDBBackend(s.App, s.Comet)
}

// PrepareV047NodeHomeConfig starts from real v2.2.1-generated documents and
// applies distinct, runtime-safe local overrides. The unique values make it
// possible to prove that config migration preserved operator intent instead of
// merely regenerating v0.50 defaults.
func PrepareV047NodeHomeConfig(app, client, comet []byte, chainID string) (NodeHomeConfigSnapshot, error) {
	chainID = strings.TrimSpace(chainID)
	if chainID == "" {
		return NodeHomeConfigSnapshot{}, errors.New("config compatibility chain ID is required")
	}
	appTree, err := toml.LoadBytes(app)
	if err != nil {
		return NodeHomeConfigSnapshot{}, fmt.Errorf("parse v0.47 app.toml: %w", err)
	}
	clientTree, err := toml.LoadBytes(client)
	if err != nil {
		return NodeHomeConfigSnapshot{}, fmt.Errorf("parse v0.47 client.toml: %w", err)
	}
	cometTree, err := toml.LoadBytes(comet)
	if err != nil {
		return NodeHomeConfigSnapshot{}, fmt.Errorf("parse v0.47 config.toml: %w", err)
	}

	// These two keys identify the v0.47 gRPC-Web shape. A freshly generated
	// v0.50 file must not be accepted as the legacy input fixture.
	for _, path := range []string{"grpc-web.address", "grpc-web.enable-unsafe-cors"} {
		if appTree.GetPath(strings.Split(path, ".")) == nil {
			return NodeHomeConfigSnapshot{}, fmt.Errorf("v0.47 app.toml is missing legacy %s", path)
		}
	}

	setTOMLPath(appTree, "app-db-backend", configCompatApplicationBackend)
	setTOMLPath(appTree, "query-gas-limit", int64(configCompatQueryGasLimit))
	setTOMLPath(appTree, "api.address", configCompatAPIListener)
	setTOMLPath(appTree, "api.enabled-unsafe-cors", false)
	setTOMLPath(appTree, "api.max-open-connections", int64(configCompatAPIMaxConnections))
	setTOMLPath(appTree, "api.rpc-max-body-bytes", int64(configCompatAPIMaxBodyBytes))
	setTOMLPath(appTree, "api.rpc-read-timeout", int64(configCompatAPIReadTimeout))
	setTOMLPath(appTree, "api.rpc-write-timeout", int64(configCompatAPIWriteTimeout))
	setTOMLPath(appTree, "grpc.max-recv-msg-size", configCompatGRPCMaxRecvMsgSize)
	setTOMLPath(appTree, "grpc.max-send-msg-size", configCompatGRPCMaxSendMsgSize)
	setTOMLPath(appTree, "grpc-web.enable", true)

	setTOMLPath(clientTree, "chain-id", chainID)
	setTOMLPath(clientTree, "node", "tcp://localhost:26657")
	setTOMLPath(clientTree, "output", "json")
	setTOMLPath(clientTree, "broadcast-mode", "sync")

	setTOMLPath(cometTree, "db_backend", configCompatCometBackend)
	setTOMLPath(cometTree, "mempool.size", int64(configCompatMempoolSize))
	setTOMLPath(cometTree, "mempool.max_txs_bytes", int64(configCompatMempoolMaxTxsBytes))
	setTOMLPath(cometTree, "p2p.laddr", configCompatP2PListener)
	setTOMLPath(cometTree, "p2p.dial_timeout", configCompatP2PDialTimeout)
	setTOMLPath(cometTree, "rpc.laddr", configCompatRPCListener)
	setTOMLPath(cometTree, "rpc.timeout_broadcast_tx_commit", configCompatRPCTxCommitTimeout)
	setTOMLPath(cometTree, "tx_index.indexer", "kv")

	prepared := NewNodeHomeConfigSnapshot(
		[]byte(appTree.String()),
		[]byte(clientTree.String()),
		[]byte(cometTree.String()),
	)
	if err := prepared.Validate(); err != nil {
		return NodeHomeConfigSnapshot{}, fmt.Errorf("validate prepared v0.47 node home: %w", err)
	}
	if err := validateConfigCompatibilityOperatorOverrides(prepared); err != nil {
		return NodeHomeConfigSnapshot{}, fmt.Errorf("validate prepared v0.47 overrides: %w", err)
	}
	return prepared, nil
}

func setTOMLPath(tree *toml.Tree, path string, value any) {
	tree.SetPath(strings.Split(path, "."), value)
}

// ValidatePreservedV047NodeHome requires that starting the current binary on
// the old volume did not opportunistically rewrite any local config file.
func ValidatePreservedV047NodeHome(before, after NodeHomeConfigSnapshot) error {
	if err := before.Validate(); err != nil {
		return fmt.Errorf("validate pre-switch node home: %w", err)
	}
	if err := validateConfigCompatibilityOperatorOverrides(before); err != nil {
		return fmt.Errorf("validate pre-switch operator overrides: %w", err)
	}
	if err := after.Validate(); err != nil {
		return fmt.Errorf("validate post-switch node home: %w", err)
	}
	for _, item := range []struct {
		name string
		want []byte
		got  []byte
	}{
		{name: "app.toml", want: before.App, got: after.App},
		{name: "client.toml", want: before.Client, got: after.Client},
		{name: "config.toml", want: before.Comet, got: after.Comet},
	} {
		if !bytes.Equal(item.want, item.got) {
			return fmt.Errorf("%s changed while starting the current binary on the v0.47 node home", item.name)
		}
	}
	return nil
}

// ValidateMigratedV050NodeHome checks the confix contract: only app.toml may
// change, every bounded endpoint/DB/operator override survives semantically,
// and the obsolete separate gRPC-Web listener keys are removed.
func ValidateMigratedV050NodeHome(before, after NodeHomeConfigSnapshot) error {
	if err := before.Validate(); err != nil {
		return fmt.Errorf("validate pre-migration node home: %w", err)
	}
	if err := validateConfigCompatibilityOperatorOverrides(before); err != nil {
		return fmt.Errorf("validate pre-migration operator overrides: %w", err)
	}
	if err := after.Validate(); err != nil {
		return fmt.Errorf("validate migrated node home: %w", err)
	}
	if !bytes.Equal(before.Client, after.Client) {
		return errors.New("client.toml changed during app.toml migration")
	}
	if !bytes.Equal(before.Comet, after.Comet) {
		return errors.New("config.toml changed during app.toml migration")
	}

	beforeApp, err := toml.LoadBytes(before.App)
	if err != nil {
		return fmt.Errorf("parse pre-migration app.toml: %w", err)
	}
	afterApp, err := toml.LoadBytes(after.App)
	if err != nil {
		return fmt.Errorf("parse migrated app.toml: %w", err)
	}
	for _, path := range []string{
		"app-db-backend",
		"query-gas-limit",
		"api.address",
		"api.enabled-unsafe-cors",
		"api.max-open-connections",
		"api.rpc-max-body-bytes",
		"api.rpc-read-timeout",
		"api.rpc-write-timeout",
		"grpc.max-recv-msg-size",
		"grpc.max-send-msg-size",
		"grpc-web.enable",
	} {
		want := beforeApp.GetPath(strings.Split(path, "."))
		got := afterApp.GetPath(strings.Split(path, "."))
		if !tomlValuesEqual(want, got) {
			return fmt.Errorf("app.toml %s changed during migration: got %v, want %v", path, got, want)
		}
	}
	for _, removed := range []string{"grpc-web.address", "grpc-web.enable-unsafe-cors"} {
		if got := afterApp.GetPath(strings.Split(removed, ".")); got != nil {
			return fmt.Errorf("migrated app.toml retained removed %s=%v", removed, got)
		}
	}
	if err := validateV050ApplicationSemanticDiff(beforeApp, afterApp); err != nil {
		return fmt.Errorf("validate app.toml v0.50 semantic diff: %w", err)
	}
	return validateBoundedApplicationConfig(afterApp)
}

func validateV050ApplicationSemanticDiff(before, after *toml.Tree) error {
	beforeValues := make(map[string]any)
	afterValues := make(map[string]any)
	flattenConfigCompatibilityTOML("", before.ToMap(), beforeValues)
	flattenConfigCompatibilityTOML("", after.ToMap(), afterValues)

	for path, beforeValue := range beforeValues {
		afterValue, present := afterValues[path]
		if !present {
			if _, allowed := configCompatV050RemovedPaths[path]; !allowed {
				return fmt.Errorf("unexpected removed path %s", path)
			}
			continue
		}
		if !tomlValuesEqual(beforeValue, afterValue) {
			return fmt.Errorf("shared path %s changed: got %v, want %v", path, afterValue, beforeValue)
		}
	}

	for path, afterValue := range afterValues {
		if _, existed := beforeValues[path]; existed {
			continue
		}
		expected, allowed := configCompatV050AddedValues[path]
		if !allowed {
			return fmt.Errorf("unexpected added path %s=%v", path, afterValue)
		}
		if !tomlValuesEqual(expected, afterValue) {
			return fmt.Errorf("added path %s=%v, want v0.50 default %v", path, afterValue, expected)
		}
	}

	for path, expected := range configCompatV050AddedValues {
		actual, present := afterValues[path]
		if !present {
			return fmt.Errorf("missing required v0.50 path %s", path)
		}
		if !tomlValuesEqual(expected, actual) {
			return fmt.Errorf("v0.50 path %s=%v, want %v", path, actual, expected)
		}
	}
	for path := range configCompatV050RemovedPaths {
		if _, existed := beforeValues[path]; !existed {
			continue
		}
		if value, retained := afterValues[path]; retained {
			return fmt.Errorf("retained removed v0.47 path %s=%v", path, value)
		}
	}
	return nil
}

func flattenConfigCompatibilityTOML(prefix string, values map[string]any, flattened map[string]any) {
	for key, value := range values {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		switch nested := value.(type) {
		case map[string]any:
			flattenConfigCompatibilityTOML(path, nested, flattened)
		default:
			flattened[path] = value
		}
	}
}

func validateBoundedApplicationConfig(app *toml.Tree) error {
	for _, path := range []string{
		"query-gas-limit",
		"api.max-open-connections",
		"api.rpc-max-body-bytes",
		"api.rpc-read-timeout",
		"api.rpc-write-timeout",
		"grpc.max-recv-msg-size",
		"grpc.max-send-msg-size",
	} {
		value := app.GetPath(strings.Split(path, "."))
		parsed, err := positiveTOMLInteger(value)
		if err != nil {
			return fmt.Errorf("app.toml %s must be a finite positive integer: %w", path, err)
		}
		if parsed <= 0 {
			return fmt.Errorf("app.toml %s must be positive, got %d", path, parsed)
		}
	}
	if got := app.GetPath([]string{"api", "address"}); fmt.Sprint(got) != configCompatAPIListener {
		return fmt.Errorf("app.toml api.address %q, want %q", got, configCompatAPIListener)
	}
	return nil
}

func validateConfigCompatibilityOperatorOverrides(snapshot NodeHomeConfigSnapshot) error {
	app, err := toml.LoadBytes(snapshot.App)
	if err != nil {
		return err
	}
	client, err := toml.LoadBytes(snapshot.Client)
	if err != nil {
		return err
	}
	comet, err := toml.LoadBytes(snapshot.Comet)
	if err != nil {
		return err
	}
	for _, expected := range []struct {
		tree *toml.Tree
		path string
		want any
	}{
		{tree: app, path: "app-db-backend", want: configCompatApplicationBackend},
		{tree: app, path: "query-gas-limit", want: int64(configCompatQueryGasLimit)},
		{tree: app, path: "api.address", want: configCompatAPIListener},
		{tree: app, path: "api.max-open-connections", want: int64(configCompatAPIMaxConnections)},
		{tree: app, path: "api.rpc-max-body-bytes", want: int64(configCompatAPIMaxBodyBytes)},
		{tree: app, path: "api.rpc-read-timeout", want: int64(configCompatAPIReadTimeout)},
		{tree: app, path: "api.rpc-write-timeout", want: int64(configCompatAPIWriteTimeout)},
		{tree: app, path: "grpc.max-recv-msg-size", want: configCompatGRPCMaxRecvMsgSize},
		{tree: app, path: "grpc.max-send-msg-size", want: configCompatGRPCMaxSendMsgSize},
		{tree: app, path: "grpc-web.enable", want: true},
		{tree: client, path: "output", want: "json"},
		{tree: client, path: "node", want: "tcp://localhost:26657"},
		{tree: client, path: "broadcast-mode", want: "sync"},
		{tree: comet, path: "db_backend", want: configCompatCometBackend},
		{tree: comet, path: "mempool.size", want: int64(configCompatMempoolSize)},
		{tree: comet, path: "mempool.max_txs_bytes", want: int64(configCompatMempoolMaxTxsBytes)},
		{tree: comet, path: "mempool.type", want: "flood"},
		{tree: comet, path: "p2p.laddr", want: configCompatP2PListener},
		{tree: comet, path: "p2p.dial_timeout", want: configCompatP2PDialTimeout},
		{tree: comet, path: "rpc.laddr", want: configCompatRPCListener},
		{tree: comet, path: "rpc.timeout_broadcast_tx_commit", want: configCompatRPCTxCommitTimeout},
		{tree: comet, path: "tx_index.indexer", want: "kv"},
	} {
		got := expected.tree.GetPath(strings.Split(expected.path, "."))
		if !tomlValuesEqual(expected.want, got) {
			return fmt.Errorf("operator override %s=%v, want %v", expected.path, got, expected.want)
		}
	}
	chainID := strings.TrimSpace(fmt.Sprint(client.Get("chain-id")))
	if chainID == "" || chainID == "<nil>" {
		return errors.New("operator override client chain-id is required")
	}
	return nil
}

func positiveTOMLInteger(value any) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case uint64:
		if typed > uint64(^uint64(0)>>1) {
			return 0, fmt.Errorf("value %d overflows int64", typed)
		}
		return int64(typed), nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported TOML type %T", value)
	}
}

func tomlValuesEqual(left, right any) bool {
	leftInteger, leftErr := positiveTOMLInteger(left)
	rightInteger, rightErr := positiveTOMLInteger(right)
	if leftErr == nil && rightErr == nil {
		return leftInteger == rightInteger
	}
	return fmt.Sprint(left) == fmt.Sprint(right)
}

func validateConfigCompatibilityDBBackend(appContents, cometContents []byte) error {
	app, err := toml.LoadBytes(appContents)
	if err != nil {
		return fmt.Errorf("parse app.toml for DB backend: %w", err)
	}
	comet, err := toml.LoadBytes(cometContents)
	if err != nil {
		return fmt.Errorf("parse config.toml for DB backend: %w", err)
	}
	appBackend := strings.ToLower(strings.TrimSpace(fmt.Sprint(app.Get("app-db-backend"))))
	cometBackend := strings.ToLower(strings.TrimSpace(fmt.Sprint(comet.Get("db_backend"))))
	for name, backend := range map[string]string{
		"app.toml app-db-backend": appBackend,
		"config.toml db_backend":  cometBackend,
	} {
		if _, unsupported := configCompatUnsupportedDBBackends[backend]; unsupported {
			return fmt.Errorf("%s uses unsupported Cosmos SDK v0.50 backend %q", name, backend)
		}
	}
	effective := appBackend
	if effective == "" {
		effective = cometBackend
	}
	if effective != configCompatCometBackend {
		return fmt.Errorf("effective application DB backend %q, want %q", effective, configCompatCometBackend)
	}
	return nil
}
