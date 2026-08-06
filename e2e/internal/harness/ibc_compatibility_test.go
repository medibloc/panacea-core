package harness

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseAndValidateIBCBinaryVersionLongUsesEffectiveDependencies(t *testing.T) {
	t.Parallel()

	contents := []byte(`build_deps:
- github.com/cometbft/cometbft@v0.37.4 => github.com/cometbft/cometbft@v0.37.18
- github.com/cosmos/cosmos-sdk@v0.47.10
- github.com/cosmos/ibc-go/v7@v7.3.2
build_tags: netgo
commit: a1b342939ba6ac3092aeebbee6a2fa741a34d47f
cosmos_sdk_version: v0.47.10
go: go version go1.26.5 linux/amd64
name: panacea-core
server_name: panacead
version: 2.2.1
`)

	identity, err := parseAndValidateIBCBinaryVersionLong(contents, panaceaV221BinaryContract())
	if err != nil {
		t.Fatalf("valid v2.2.1 version --long rejected: %v", err)
	}
	if identity.Version != "2.2.1" || identity.Commit != panaceaV221SourceCommit {
		t.Fatalf("identity = %#v", identity)
	}
	if got := identity.DependencyVersion(cometBFTModulePath); got != "v0.37.18" {
		t.Fatalf("effective CometBFT dependency = %q, want v0.37.18", got)
	}
	if got := identity.DependencyVersion(ibcGoV7ModulePath); got != "v7.3.2" {
		t.Fatalf("IBC-Go dependency = %q, want v7.3.2", got)
	}

	wrong := strings.Replace(string(contents), "github.com/cosmos/ibc-go/v7@v7.3.2", "github.com/cosmos/ibc-go/v7@v7.3.1", 1)
	if _, err := parseAndValidateIBCBinaryVersionLong([]byte(wrong), panaceaV221BinaryContract()); err == nil {
		t.Fatal("wrong IBC-Go dependency unexpectedly accepted")
	}
}

func TestParseAndValidateOsmosisVersionLongPreservesCrossModuleReplacement(t *testing.T) {
	t.Parallel()

	contents := []byte(`build_deps:
- github.com/cometbft/cometbft@v0.38.22
- github.com/cosmos/cosmos-sdk@v0.50.14 => github.com/osmosis-labs/cosmos-sdk@v0.50.14-v30-osmo
- github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8@v8.2.0
- github.com/cosmos/ibc-go/v8@v8.7.0
- github.com/osmosis-labs/osmosis/x/ibc-hooks@v0.0.21
build_tags: netgo,ledger,muslc
commit: a56c05b0e83341b9a3c0e6e3508520f15e9f2e49
cosmos_sdk_version: v0.50.14-v30-osmo
go: go version go1.23.4 linux/amd64
name: osmosis
server_name: osmosisd
version: 31.0.2
`)

	identity, err := parseAndValidateIBCBinaryVersionLong(contents, pinnedOsmosisBinaryContract())
	if err != nil {
		t.Fatalf("valid Osmosis replacement rejected: %v", err)
	}
	sdkDependency, ok := identity.Dependency(cosmosSDKModulePath)
	if !ok {
		t.Fatal("original Cosmos SDK module path was discarded")
	}
	if sdkDependency.Version != "v0.50.14" || sdkDependency.Replacement == nil ||
		sdkDependency.Replacement.Path != "github.com/osmosis-labs/cosmos-sdk" ||
		sdkDependency.Replacement.Version != "v0.50.14-v30-osmo" {
		t.Fatalf("Cosmos SDK replacement = %#v", sdkDependency)
	}
	if _, ok := identity.Dependency("github.com/osmosis-labs/cosmos-sdk"); ok {
		t.Fatal("replacement module was incorrectly indexed as the original dependency")
	}

	wrongReplacementPath := strings.Replace(
		string(contents),
		"github.com/osmosis-labs/cosmos-sdk@v0.50.14-v30-osmo",
		"github.com/example/cosmos-sdk@v0.50.14-v30-osmo",
		1,
	)
	if _, err := parseAndValidateIBCBinaryVersionLong([]byte(wrongReplacementPath), pinnedOsmosisBinaryContract()); err == nil {
		t.Fatal("wrong Cosmos SDK replacement path unexpectedly accepted")
	}
	wrongReplacementVersion := strings.Replace(string(contents), "v0.50.14-v30-osmo", "v0.50.14-v29-osmo", 2)
	if _, err := parseAndValidateIBCBinaryVersionLong([]byte(wrongReplacementVersion), pinnedOsmosisBinaryContract()); err == nil {
		t.Fatal("wrong Cosmos SDK replacement version unexpectedly accepted")
	}
}

func TestOsmosisMiddlewareEvidenceDoesNotClaimLiveWiringConfirmation(t *testing.T) {
	t.Parallel()

	evidence := expectedOsmosisMiddlewareEvidence()
	if evidence.InvestigationStatus == ibcInvestigationConfirmed {
		t.Fatal("pinned source wiring must not be reported as live-confirmed middleware")
	}
	if evidence.InvestigationStatus != "pinned-source-live-limited" {
		t.Fatalf("middleware investigation status = %q", evidence.InvestigationStatus)
	}
	if evidence.WiringObservedLive || evidence.PerChannelActivationObservedLive {
		t.Fatalf("middleware evidence overstates live visibility: %#v", evidence)
	}
	if err := evidence.Validate(); err != nil {
		t.Fatalf("limited middleware evidence rejected: %v", err)
	}

	overstated := evidence
	overstated.WiringObservedLive = true
	if err := overstated.Validate(); err == nil {
		t.Fatal("live wiring overstatement unexpectedly accepted")
	}
}

func TestPinnedOsmosisSourceContractRecordsReplacementAndWiringProvenance(t *testing.T) {
	t.Parallel()

	contract := pinnedOsmosisSourceContract()
	if err := contract.Validate(); err != nil {
		t.Fatalf("pinned source contract rejected: %v", err)
	}
	if contract.SDKDependency.Path != cosmosSDKModulePath || contract.SDKDependency.Version != "v0.50.14" ||
		contract.SDKDependency.Replacement == nil || contract.SDKDependency.Replacement.Path != osmosisSDKModulePath ||
		contract.SDKDependency.Replacement.Version != "v0.50.14-v30-osmo" {
		t.Fatalf("pinned SDK replacement = %#v", contract.SDKDependency)
	}
	if !strings.Contains(contract.GoMod.Reference, osmosisSourceCommit) ||
		!strings.Contains(contract.TransferWiring.Reference, osmosisSourceCommit) {
		t.Fatalf("source references are not commit-pinned: %#v", contract)
	}
	if contract.GoMod.SHA256 != osmosisGoModSHA256 || contract.TransferWiring.SHA256 != osmosisTransferWiringSHA256 {
		t.Fatalf("source SHA-256 contracts = %#v %#v", contract.GoMod, contract.TransferWiring)
	}

	tampered := contract
	tampered.SDKDependency.Replacement = &IBCDependencyReplacement{Path: osmosisSDKModulePath, Version: "v0.50.14-v29-osmo"}
	if err := tampered.Validate(); err == nil {
		t.Fatal("mutated source replacement unexpectedly accepted")
	}
}

func TestOsmosisMainnetFixturePinsRegistryReleaseAndEndpoints(t *testing.T) {
	t.Setenv(osmosisMainnetRPCEnv, "https://rpc.operator.example")
	t.Setenv(osmosisMainnetRESTEnv, "https://rest.operator.example")

	config, err := resolveOsmosisMainnetPreflightConfig()
	if err != nil {
		t.Fatalf("resolve preflight config: %v", err)
	}
	if config.RPCEndpoint != "https://rpc.operator.example" || config.RESTEndpoint != "https://rest.operator.example" {
		t.Fatalf("explicit endpoints were not retained: %#v", config)
	}
	fixture := expectedOsmosisMainnetFixture()
	if fixture.ChainID != "osmosis-1" || fixture.CounterpartyChainID != "panacea-3" ||
		fixture.ChannelID != "channel-82" || fixture.CounterpartyChannelID != "channel-1" ||
		fixture.ConnectionID != "connection-1231" || fixture.Version != "ics20-1" ||
		fixture.RegistryStatus != "ACTIVE" || !fixture.RegistryPreferred {
		t.Fatalf("mainnet fixture = %#v", fixture)
	}
	if !strings.Contains(fixture.ChannelRegistryReference, osmosisMainnetRegistryCommit) {
		t.Fatalf("registry fixture is not commit-pinned: %q", fixture.ChannelRegistryReference)
	}
	middleware := expectedOsmosisMiddlewareEvidence()
	if err := middleware.Validate(); err != nil {
		t.Fatalf("pinned middleware contract rejected: %v", err)
	}
}

func TestCurrentPanaceaBinaryContractRequiresExpectedBuildIdentity(t *testing.T) {
	t.Setenv("PANACEA_E2E_CURRENT_BINARY_VERSION", "2.3.0")
	t.Setenv("PANACEA_E2E_CURRENT_COMMIT", strings.Repeat("b", 40))

	contract, err := currentPanaceaBinaryContract()
	if err != nil {
		t.Fatalf("currentPanaceaBinaryContract: %v", err)
	}
	if contract.Version != "2.3.0" || contract.Commit != strings.Repeat("b", 40) {
		t.Fatalf("contract = %#v", contract)
	}
	if got := contract.DependencyVersion(ibcGoV8ModulePath); got != "v8.8.0" {
		t.Fatalf("current IBC-Go contract = %q, want v8.8.0", got)
	}

	t.Setenv("PANACEA_E2E_CURRENT_COMMIT", "")
	if _, err := currentPanaceaBinaryContract(); err == nil {
		t.Fatal("missing current commit unexpectedly accepted")
	}
}

func TestParseBinaryChecksumOutputRequiresExactBinaryPath(t *testing.T) {
	t.Parallel()

	checksum := strings.Repeat("a", 64)
	gotChecksum, gotPath, err := parseBinaryChecksumOutput(
		[]byte(checksum+"  /usr/bin/panacead\n"),
		"panacead",
	)
	if err != nil {
		t.Fatalf("valid checksum rejected: %v", err)
	}
	if gotChecksum != checksum || gotPath != "/usr/bin/panacead" {
		t.Fatalf("checksum parse = %q %q", gotChecksum, gotPath)
	}
	for _, invalid := range []string{
		strings.Repeat("A", 64) + "  /usr/bin/panacead",
		strings.Repeat("a", 63) + "  /usr/bin/panacead",
		strings.Repeat("a", 64) + "  panacead",
		strings.Repeat("a", 64) + "  /usr/bin/osmosisd",
	} {
		if _, _, err := parseBinaryChecksumOutput([]byte(invalid), "panacead"); err == nil {
			t.Fatalf("invalid checksum output %q unexpectedly accepted", invalid)
		}
	}
}

func TestChecksumIBCContainerBinaryArchiveSupportsDistrolessImage(t *testing.T) {
	t.Parallel()

	binary := []byte("fixed-osmosisd-binary")
	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	if err := writer.WriteHeader(&tar.Header{Name: "osmosisd", Mode: 0o755, Size: int64(len(binary)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := writer.Write(binary); err != nil {
		t.Fatalf("write tar binary: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}

	checksum, err := checksumIBCContainerBinaryArchive(bytes.NewReader(archive.Bytes()), "/bin/osmosisd")
	if err != nil {
		t.Fatalf("checksum distroless binary archive: %v", err)
	}
	want := sha256.Sum256(binary)
	if checksum != hex.EncodeToString(want[:]) {
		t.Fatalf("checksum = %q, want %x", checksum, want)
	}
}

func TestHermesUpgradeInvarianceUsesLiveConfigAndRejectsRuntimeChanges(t *testing.T) {
	t.Parallel()

	identity, err := validateHermesRuntimeIdentity(
		"hermes 1.8.2",
		strings.Repeat("e", 64)+"  /usr/local/bin/hermes",
		PinnedIBCProvenance(),
	)
	if err != nil {
		t.Fatalf("Hermes identity fixture: %v", err)
	}
	chainIDs := []string{"panacea-local", "osmosis-local"}
	config := []byte(`
[[chains]]
id = "panacea-local"
compat_mode = "0.37"

[[chains]]
id = "osmosis-local"
compat_mode = "0.37"
`)

	evidence, err := buildHermesUpgradeInvarianceEvidence(identity, identity, config, config, chainIDs)
	if err != nil {
		t.Fatalf("unchanged Hermes runtime rejected: %v", err)
	}
	if !evidence.Validated || !evidence.BinaryUnchanged || !evidence.ConfigUnchanged {
		t.Fatalf("invariance evidence = %#v", evidence)
	}
	if evidence.CompatMode["panacea-local"] != "0.37" || evidence.CompatMode["osmosis-local"] != "0.37" {
		t.Fatalf("compat_mode was not parsed from live config: %#v", evidence.CompatMode)
	}

	changedBinary := identity
	changedBinary.BinarySHA256 = strings.Repeat("f", 64)
	changedEvidence, err := buildHermesUpgradeInvarianceEvidence(identity, changedBinary, config, config, chainIDs)
	if err == nil || changedEvidence.BinaryUnchanged || changedEvidence.Validated {
		t.Fatalf("changed Hermes binary accepted: evidence=%#v err=%v", changedEvidence, err)
	}

	changedConfig := append(append([]byte(nil), config...), []byte("\n[telemetry]\nenabled = true\n")...)
	changedEvidence, err = buildHermesUpgradeInvarianceEvidence(identity, identity, config, changedConfig, chainIDs)
	if err == nil || changedEvidence.ConfigUnchanged || changedEvidence.Validated {
		t.Fatalf("changed Hermes config accepted: evidence=%#v err=%v", changedEvidence, err)
	}

	wrongCompat := bytes.ReplaceAll(config, []byte(`compat_mode = "0.37"`), []byte(`compat_mode = "0.38"`))
	changedEvidence, err = buildHermesUpgradeInvarianceEvidence(identity, identity, wrongCompat, wrongCompat, chainIDs)
	if err == nil || changedEvidence.CompatMode["panacea-local"] != "0.38" {
		t.Fatalf("live compat_mode mismatch was hidden: evidence=%#v err=%v", changedEvidence, err)
	}
}

func TestOsmosisMainnetPreflightConfirmsActiveChannelAndLimitsMiddlewareClaim(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 4, 11, 5, 0, 0, time.UTC)
	client := &http.Client{Transport: fixtureRoundTripper{
		responses: map[string]string{
			"https://rpc.example/status":                                                  osmosisStatusFixture(now.Add(-5 * time.Second)),
			"https://rest.example/cosmos/base/tendermint/v1beta1/node_info":               osmosisNodeInfoFixture(),
			"https://rest.example/ibc/core/channel/v1/channels/channel-82/ports/transfer": osmosisChannelFixture("channel-1"),
		},
	}}

	evidence, err := runOsmosisMainnetPreflight(
		context.Background(),
		client,
		OsmosisMainnetPreflightConfig{
			RPCEndpoint:  "https://rpc.example",
			RESTEndpoint: "https://rest.example",
			Timeout:      time.Second,
		},
		now,
	)
	if err != nil {
		t.Fatalf("valid mainnet preflight rejected: %v", err)
	}
	if !evidence.Passed || evidence.Channel.InvestigationStatus != ibcInvestigationConfirmed {
		t.Fatalf("preflight evidence = %#v", evidence)
	}
	if evidence.Channel.ChannelID != "channel-82" || evidence.Channel.CounterpartyChannelID != "channel-1" {
		t.Fatalf("channel evidence = %#v", evidence.Channel)
	}
	if evidence.Middleware.InvestigationStatus != ibcInvestigationPinnedSourceLiveLimited || len(evidence.Middleware.RecvStack) != 4 ||
		evidence.Middleware.WiringObservedLive || evidence.Middleware.PerChannelActivationObservedLive {
		t.Fatalf("middleware evidence = %#v", evidence.Middleware)
	}
	if evidence.NodeInfo.DependencyObservationScope != osmosisNodeInfoDependencyScope || evidence.NodeInfo.ReplacementMetadataObserved {
		t.Fatalf("node-info observation scope = %#v", evidence.NodeInfo)
	}
}

func TestOsmosisMainnetPreflightRetriesTransientServiceUnavailable(t *testing.T) {
	now := time.Date(2026, 8, 4, 11, 5, 0, 0, time.UTC)
	transport := &transientOsmosisStatusRoundTripper{now: now}
	client := &http.Client{Transport: transport}

	evidence, err := runOsmosisMainnetPreflight(
		context.Background(),
		client,
		OsmosisMainnetPreflightConfig{
			RPCEndpoint:  "https://rpc.example",
			RESTEndpoint: "https://rest.example",
			Timeout:      5 * time.Second,
		},
		now,
	)
	if err != nil {
		t.Fatalf("transient 503 was not recovered by a bounded retry: %v", err)
	}
	if !evidence.Passed {
		t.Fatalf("preflight did not pass after transient recovery: %#v", evidence)
	}
	if transport.statusAttempts != 2 {
		t.Fatalf("status attempts = %d, want 2", transport.statusAttempts)
	}
}

func TestOsmosisMainnetPreflightFailsClosedWithArtifactEvidence(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 4, 11, 5, 0, 0, time.UTC)
	t.Run("fixture mismatch", func(t *testing.T) {
		client := &http.Client{Transport: fixtureRoundTripper{responses: map[string]string{
			"https://rpc.example/status":                                                  osmosisStatusFixture(now.Add(-5 * time.Second)),
			"https://rest.example/cosmos/base/tendermint/v1beta1/node_info":               osmosisNodeInfoFixture(),
			"https://rest.example/ibc/core/channel/v1/channels/channel-82/ports/transfer": osmosisChannelFixture("channel-999"),
		}}}
		evidence, err := runOsmosisMainnetPreflight(context.Background(), client, OsmosisMainnetPreflightConfig{
			RPCEndpoint: "https://rpc.example", RESTEndpoint: "https://rest.example", Timeout: time.Second,
		}, now)
		if err == nil || evidence.Passed {
			t.Fatalf("mismatched fixture accepted: evidence=%#v err=%v", evidence, err)
		}
		if evidence.Error == "" || evidence.Channel.InvestigationStatus != ibcInvestigationMismatch {
			t.Fatalf("mismatch was not retained in evidence: %#v", evidence)
		}
	})

	t.Run("endpoint unavailable", func(t *testing.T) {
		client := &http.Client{Transport: fixtureRoundTripper{err: errors.New("firewall denied")}}
		evidence, err := runOsmosisMainnetPreflight(context.Background(), client, OsmosisMainnetPreflightConfig{
			RPCEndpoint: "https://rpc.example", RESTEndpoint: "https://rest.example", Timeout: time.Second,
		}, now)
		if err == nil || evidence.Passed {
			t.Fatalf("unavailable endpoint accepted: evidence=%#v err=%v", evidence, err)
		}
		if evidence.Channel.InvestigationStatus != ibcInvestigationUnavailable || evidence.Error == "" {
			t.Fatalf("unavailability was not retained in evidence: %#v", evidence)
		}
	})

	t.Run("binary dependency mismatch", func(t *testing.T) {
		wrongNodeInfo := strings.Replace(osmosisNodeInfoFixture(), `"version":"v8.7.0"`, `"version":"v8.6.0"`, 1)
		client := &http.Client{Transport: fixtureRoundTripper{responses: map[string]string{
			"https://rpc.example/status":                                                  osmosisStatusFixture(now.Add(-5 * time.Second)),
			"https://rest.example/cosmos/base/tendermint/v1beta1/node_info":               wrongNodeInfo,
			"https://rest.example/ibc/core/channel/v1/channels/channel-82/ports/transfer": osmosisChannelFixture("channel-1"),
		}}}
		evidence, err := runOsmosisMainnetPreflight(context.Background(), client, OsmosisMainnetPreflightConfig{
			RPCEndpoint: "https://rpc.example", RESTEndpoint: "https://rest.example", Timeout: time.Second,
		}, now)
		if err == nil || evidence.Passed || evidence.NodeInfo.Error == "" {
			t.Fatalf("wrong live dependency accepted: evidence=%#v err=%v", evidence, err)
		}
		if evidence.Middleware.InvestigationStatus != ibcInvestigationMismatch {
			t.Fatalf("middleware mismatch status = %q", evidence.Middleware.InvestigationStatus)
		}
	})
}

func TestIBCCompatibilityMatrixRejectsMutableRuntimeEvidence(t *testing.T) {
	t.Setenv("PANACEA_E2E_CURRENT_BINARY_VERSION", "2.3.0")
	t.Setenv("PANACEA_E2E_CURRENT_COMMIT", strings.Repeat("b", 40))

	matrix := validIBCCompatibilityMatrixFixture(t)
	if err := matrix.Validate(); err != nil {
		t.Fatalf("valid compatibility matrix rejected: %v", err)
	}

	tampered := matrix
	tampered.Osmosis.PostUpgrade.Nodes = append([]IBCBinaryNodeEvidence(nil), matrix.Osmosis.PostUpgrade.Nodes...)
	tampered.Osmosis.PostUpgrade.Nodes[0].BinarySHA256 = strings.Repeat("f", 64)
	if err := tampered.Validate(); err == nil {
		t.Fatal("changed Osmosis binary checksum unexpectedly accepted")
	}

	tampered = matrix
	tampered.Osmosis.Genesis.PostUpgrade.Nodes = append([]IBCGenesisNodeChecksum(nil), matrix.Osmosis.Genesis.PostUpgrade.Nodes...)
	tampered.Osmosis.Genesis.PostUpgrade.Nodes[0].SHA256 = strings.Repeat("e", 64)
	if err := tampered.Validate(); err == nil {
		t.Fatal("changed Osmosis genesis checksum unexpectedly accepted")
	}

	tampered = matrix
	tampered.Channel.Version = "ics20-2"
	if err := tampered.Validate(); err == nil {
		t.Fatal("changed ICS-20 app version unexpectedly accepted")
	}
}

func validIBCCompatibilityMatrixFixture(t *testing.T) IBCCompatibilityMatrix {
	t.Helper()
	now := time.Date(2026, 8, 4, 11, 5, 0, 0, time.UTC)
	client := &http.Client{Transport: fixtureRoundTripper{responses: map[string]string{
		"https://rpc.example/status":                                                  osmosisStatusFixture(now.Add(-5 * time.Second)),
		"https://rest.example/cosmos/base/tendermint/v1beta1/node_info":               osmosisNodeInfoFixture(),
		"https://rest.example/ibc/core/channel/v1/channels/channel-82/ports/transfer": osmosisChannelFixture("channel-1"),
	}}}
	preflight, err := runOsmosisMainnetPreflight(context.Background(), client, OsmosisMainnetPreflightConfig{
		RPCEndpoint: "https://rpc.example", RESTEndpoint: "https://rest.example", Timeout: time.Second,
	}, now)
	if err != nil {
		t.Fatalf("build mainnet fixture: %v", err)
	}
	current, err := currentPanaceaBinaryContract()
	if err != nil {
		t.Fatalf("current contract: %v", err)
	}
	chainEvidence := func(chainID, phase, path string, contract IBCBinaryVersionContract, checksumByte string) IBCChainBinaryEvidence {
		dependencies := append([]IBCDependencyContract(nil), contract.Dependencies...)
		return IBCChainBinaryEvidence{
			ChainID:  chainID,
			Phase:    phase,
			Contract: contract,
			Nodes: []IBCBinaryNodeEvidence{{
				Name: "node-0",
				Version: IBCBinaryVersionIdentity{
					Name: contract.Name, AppName: contract.AppName, Version: contract.Version,
					Commit: contract.Commit, CosmosSDKVersion: contract.CosmosSDKVersion,
					GoVersion: "go version go1.26.5 linux/amd64", Dependencies: dependencies,
				},
				BinaryPath: path, BinarySHA256: strings.Repeat(checksumByte, 64),
			}},
			Validated: true,
		}
	}
	panaceaPre := chainEvidence("panacea-local", "pre-upgrade", "/usr/bin/panacead", panaceaV221BinaryContract(), "a")
	panaceaPost := chainEvidence("panacea-local", "post-upgrade", "/usr/bin/panacead", current, "b")
	osmosisPre := chainEvidence("osmosis-local", "pre-upgrade", "/bin/osmosisd", pinnedOsmosisBinaryContract(), "c")
	osmosisPost := chainEvidence("osmosis-local", "post-upgrade", "/bin/osmosisd", pinnedOsmosisBinaryContract(), "c")
	genesisChecksum := strings.Repeat("d", 64)
	genesis := IBCGenesisImmutabilityEvidence{
		ChainID: "osmosis-local", File: "config/genesis.json", Immutable: true,
		Initial:     IBCGenesisChecksumSnapshot{Phase: "pre-upgrade", Common: genesisChecksum, Nodes: []IBCGenesisNodeChecksum{{Name: "node-0", SHA256: genesisChecksum}}},
		PostUpgrade: IBCGenesisChecksumSnapshot{Phase: "post-upgrade", Common: genesisChecksum, Nodes: []IBCGenesisNodeChecksum{{Name: "node-0", SHA256: genesisChecksum}}},
	}
	hermesRuntime, err := validateHermesRuntimeIdentity(
		"hermes 1.8.2",
		strings.Repeat("e", 64)+"  /usr/local/bin/hermes",
		PinnedIBCProvenance(),
	)
	if err != nil {
		t.Fatalf("Hermes fixture: %v", err)
	}
	return IBCCompatibilityMatrix{
		SchemaVersion:    ibcCompatibilityMatrixSchema,
		GeneratedAt:      now,
		MainnetPreflight: preflight,
		Panacea:          IBCBinaryUpgradeMatrix{PreUpgrade: panaceaPre, PostUpgrade: panaceaPost},
		Osmosis: IBCOsmosisCompatibilityMatrix{
			SourceContract: pinnedOsmosisSourceContract(),
			PreUpgrade:     osmosisPre,
			PostUpgrade:    osmosisPost,
			Genesis:        genesis,
		},
		Channel: IBCCompatibilityChannelContract{
			PanaceaChannelID: "channel-0", OsmosisChannelID: "channel-0", PortID: "transfer",
			State: "STATE_OPEN", Ordering: "ORDER_UNORDERED", Version: "ics20-1",
		},
		Middleware: expectedOsmosisMiddlewareEvidence(),
		Hermes: IBCHermesCompatibilityMatrix{
			RuntimeIdentity: hermesRuntime,
			CompatMode:      map[string]string{"panacea-local": "0.37", "osmosis-local": "0.37"},
		},
		Validated: true,
	}
}

type fixtureRoundTripper struct {
	responses map[string]string
	err       error
}

type transientOsmosisStatusRoundTripper struct {
	now            time.Time
	statusAttempts int
}

func (t *transientOsmosisStatusRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body string
	switch req.URL.String() {
	case "https://rpc.example/status":
		t.statusAttempts++
		if t.statusAttempts == 1 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Status:     "503 Service Unavailable",
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("temporarily unavailable")),
				Request:    req,
			}, nil
		}
		body = osmosisStatusFixture(t.now.Add(-5 * time.Second))
	case "https://rest.example/cosmos/base/tendermint/v1beta1/node_info":
		body = osmosisNodeInfoFixture()
	case "https://rest.example/ibc/core/channel/v1/channels/channel-82/ports/transfer":
		body = osmosisChannelFixture("channel-1")
	default:
		return nil, errors.New("unexpected URL: " + req.URL.String())
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func (f fixtureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	body, ok := f.responses[req.URL.String()]
	if !ok {
		return nil, errors.New("unexpected URL: " + req.URL.String())
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func osmosisStatusFixture(blockTime time.Time) string {
	return `{"jsonrpc":"2.0","result":{"node_info":{"network":"osmosis-1","version":"0.38.22"},"sync_info":{"latest_block_height":"67805144","latest_block_time":"` + blockTime.Format(time.RFC3339Nano) + `","catching_up":false}}}`
}

func osmosisNodeInfoFixture() string {
	return `{
  "default_node_info":{"network":"osmosis-1","version":"0.38.22"},
  "application_version":{
    "name":"osmosis","app_name":"osmosisd","version":"31.0.2",
    "git_commit":"a56c05b0e83341b9a3c0e6e3508520f15e9f2e49",
    "go_version":"go version go1.23.4 linux/amd64",
    "cosmos_sdk_version":"v0.50.14-v30-osmo",
    "build_deps":[
      {"path":"github.com/cometbft/cometbft","version":"v0.38.22"},
      {"path":"github.com/cosmos/cosmos-sdk","version":"v0.50.14"},
      {"path":"github.com/cosmos/ibc-go/v8","version":"v8.7.0"},
      {"path":"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8","version":"v8.2.0"},
      {"path":"github.com/osmosis-labs/osmosis/x/ibc-hooks","version":"v0.0.21"}
    ]
  }
}`
}

func osmosisChannelFixture(counterpartyChannel string) string {
	return `{"channel":{"state":"STATE_OPEN","ordering":"ORDER_UNORDERED","counterparty":{"port_id":"transfer","channel_id":"` + counterpartyChannel + `"},"connection_hops":["connection-1231"],"version":"ics20-1"},"proof_height":{"revision_number":"1","revision_height":"67805146"}}`
}
