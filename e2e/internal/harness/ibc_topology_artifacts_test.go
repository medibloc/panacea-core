package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
)

func TestHermesRuntimeIdentityPinsVersionSourceImageAndBinaryChecksum(t *testing.T) {
	checksum := strings.Repeat("a", 64) + "  /usr/local/bin/hermes\n"
	identity, err := validateHermesRuntimeIdentity("hermes 1.8.2\n", checksum, PinnedIBCProvenance())
	if err != nil {
		t.Fatalf("valid Hermes runtime identity rejected: %v", err)
	}
	if identity.ReleaseIdentifier != "1.8.2+06dfbaf" {
		t.Fatalf("release identifier = %q, want 1.8.2+06dfbaf", identity.ReleaseIdentifier)
	}
	if identity.VersionOutput != "hermes 1.8.2" {
		t.Fatalf("version output = %q, want hermes 1.8.2", identity.VersionOutput)
	}
	if identity.BinarySHA256 != strings.Repeat("a", 64) || identity.BinaryPath != "/usr/local/bin/hermes" {
		t.Fatalf("binary identity = %s %s", identity.BinarySHA256, identity.BinaryPath)
	}
	if identity.ImageReference != PinnedHermesImage().Ref() || identity.SourceCommit != hermesSourceCommit {
		t.Fatalf("source/image identity is not pinned: %#v", identity)
	}

	for _, version := range []string{"hermes 1.8.1", "hermes 1.8.2+06dfbaf", "1.8.2"} {
		if _, err := validateHermesRuntimeIdentity(version, checksum, PinnedIBCProvenance()); err == nil {
			t.Fatalf("non-exact Hermes version %q unexpectedly accepted", version)
		}
	}
	if _, err := validateHermesRuntimeIdentity("hermes 1.8.2", "not-a-checksum  /usr/local/bin/hermes", PinnedIBCProvenance()); err == nil {
		t.Fatal("invalid Hermes binary checksum unexpectedly accepted")
	}

	wrongSource := PinnedIBCProvenance()
	wrongSource.Hermes.SourceCommit = strings.Repeat("0", 40)
	if _, err := validateHermesRuntimeIdentity("hermes 1.8.2", checksum, wrongSource); err == nil {
		t.Fatal("wrong Hermes source commit unexpectedly accepted")
	}
	wrongImage := PinnedIBCProvenance()
	wrongImage.Hermes.Reference = "ghcr.io/informalsystems/hermes:1.8.2"
	if _, err := validateHermesRuntimeIdentity("hermes 1.8.2", checksum, wrongImage); err == nil {
		t.Fatal("unpinned Hermes image unexpectedly accepted")
	}
}

func TestIBCTopologyRecordTestPanicDelegatesToBaseArtifactsAndRepanics(t *testing.T) {
	store, err := newArtifactStore(
		"ibc panic",
		"ibc-panic-a1",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "ibc-panic-artifacts")),
	)
	if err != nil {
		t.Fatalf("newArtifactStore: %v", err)
	}
	topology := &IBCTopology{artifacts: &ibcTopologyArtifacts{base: store}}

	recovered := invokeIBCTopologyTestPanicRecorder(topology, "ibc packet reflection panic")
	if recovered != "ibc packet reflection panic" {
		t.Fatalf("recovered panic = %#v, want original panic value", recovered)
	}
	store.mu.Lock()
	failed := store.failed
	state := store.state
	failures := append([]artifactFailure(nil), store.failures...)
	store.mu.Unlock()
	if !failed || state != "failed" {
		t.Fatalf("IBC panic artifact state = failed:%t state:%q", failed, state)
	}
	if len(failures) != 1 || failures[0].Stage != "test-panic" ||
		!strings.Contains(failures[0].Error, "ibc packet reflection panic") {
		t.Fatalf("IBC panic failures = %#v", failures)
	}
}

func invokeIBCTopologyTestPanicRecorder(topology *IBCTopology, panicValue any) (recovered any) {
	defer func() {
		recovered = recover()
	}()
	func() {
		defer topology.RecordTestPanic()
		panic(panicValue)
	}()
	return nil
}

func TestIBCTopologyArtifactsAreOwnedByOneRun(t *testing.T) {
	root := filepath.Join(trustedArtifactTempDir(t), "ibc-artifacts")
	base, err := newArtifactStore("ibc smoke", "ibc-artifacts-a1", artifactTestConfig(root))
	if err != nil {
		t.Fatalf("newArtifactStore: %v", err)
	}

	descriptor := IBCTopologyDescriptor{
		Path:              "panacea-osmosis",
		PanaceaChainID:    "panacea-ibc-artifacts-a1",
		OsmosisChainID:    "osmosis-ibc-artifacts-a1",
		PanaceaValidators: 1,
		PanaceaFullNodes:  1,
		OsmosisValidators: 1,
		OsmosisFullNodes:  1,
	}
	artifacts, err := newIBCTopologyArtifacts(base, descriptor)
	if err != nil {
		t.Fatalf("newIBCTopologyArtifacts: %v", err)
	}
	if err := artifacts.closeReporter(); err != nil {
		t.Fatalf("closeReporter: %v", err)
	}
	if err := artifacts.closeReporter(); err != nil {
		t.Fatalf("idempotent closeReporter: %v", err)
	}

	provenanceRaw, err := os.ReadFile(filepath.Join(base.dir, "ibc", "provenance.json"))
	if err != nil {
		t.Fatalf("read provenance: %v", err)
	}
	var provenance IBCReleaseProvenance
	if err := json.Unmarshal(provenanceRaw, &provenance); err != nil {
		t.Fatalf("decode provenance: %v", err)
	}
	if got, want := provenance.Hermes.Reference, PinnedHermesImage().Ref(); got != want {
		t.Fatalf("Hermes reference = %q, want %q", got, want)
	}

	sourceRaw, err := os.ReadFile(filepath.Join(base.dir, OsmosisPinnedSourceContractArtifactPath))
	if err != nil {
		t.Fatalf("read Osmosis source contract: %v", err)
	}
	var sourceContract OsmosisPinnedSourceContract
	if err := json.Unmarshal(sourceRaw, &sourceContract); err != nil {
		t.Fatalf("decode Osmosis source contract: %v", err)
	}
	if err := sourceContract.Validate(); err != nil {
		t.Fatalf("validate recorded Osmosis source contract: %v", err)
	}

	topologyRaw, err := os.ReadFile(filepath.Join(base.dir, "ibc", "topology.json"))
	if err != nil {
		t.Fatalf("read topology: %v", err)
	}
	var recorded IBCTopologyDescriptor
	if err := json.Unmarshal(topologyRaw, &recorded); err != nil {
		t.Fatalf("decode topology: %v", err)
	}
	if recorded != descriptor {
		t.Fatalf("recorded topology = %#v, want %#v", recorded, descriptor)
	}

	reporterPath := filepath.Join(base.dir, "ibc", "hermes", "exec-report.jsonl")
	reporterInfo, err := os.Stat(reporterPath)
	if err != nil {
		t.Fatalf("stat reporter: %v", err)
	}
	if got := reporterInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("reporter mode = %o, want 600", got)
	}
	reportRaw, err := os.ReadFile(reporterPath)
	if err != nil {
		t.Fatalf("read reporter: %v", err)
	}
	if report := string(reportRaw); !strings.Contains(report, `"Type":"BeginSuite"`) || !strings.Contains(report, `"Type":"FinishSuite"`) {
		t.Fatalf("reporter does not contain a complete suite lifecycle: %s", report)
	}
}

func TestPinHermesCompatModeCoversBothLocalChains(t *testing.T) {
	input := []byte(`
[[chains]]
id = "panacea-ibc-config-a1"
rpc_addr = "http://panacea:26657"

[[chains]]
id = "osmosis-ibc-config-a1"
rpc_addr = "http://osmosis:26657"
`)

	output, err := pinHermesCompatMode(input, []string{
		"panacea-ibc-config-a1",
		"osmosis-ibc-config-a1",
	})
	if err != nil {
		t.Fatalf("pinHermesCompatMode: %v", err)
	}

	tree, err := toml.LoadBytes(output)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	chains, ok := tree.Get("chains").([]*toml.Tree)
	if !ok || len(chains) != 2 {
		t.Fatalf("chains = %#v, want two TOML tables", tree.Get("chains"))
	}
	for _, chain := range chains {
		if got := chain.Get("compat_mode"); got != "0.37" {
			t.Fatalf("chain %v compat_mode = %#v, want 0.37", chain.Get("id"), got)
		}
	}
}

func TestPinHermesPacketFiltersAllowsOnlyTheHandshakeChannels(t *testing.T) {
	input := []byte(`
[[chains]]
id = "panacea-ibc-filter-a1"
compat_mode = "0.37"

[[chains]]
id = "osmosis-ibc-filter-a1"
compat_mode = "0.37"
`)

	output, err := pinHermesPacketFilters(input, map[string]string{
		"panacea-ibc-filter-a1": "channel-2",
		"osmosis-ibc-filter-a1": "channel-7",
	})
	if err != nil {
		t.Fatalf("pinHermesPacketFilters: %v", err)
	}

	var decoded struct {
		Chains []struct {
			ID           string `toml:"id"`
			CompatMode   string `toml:"compat_mode"`
			PacketFilter struct {
				Policy string     `toml:"policy"`
				List   [][]string `toml:"list"`
			} `toml:"packet_filter"`
		} `toml:"chains"`
	}
	if err := toml.Unmarshal(output, &decoded); err != nil {
		t.Fatalf("decode filtered config: %v", err)
	}
	if len(decoded.Chains) != 2 {
		t.Fatalf("decoded %d chains, want 2", len(decoded.Chains))
	}
	wantChannels := map[string]string{
		"panacea-ibc-filter-a1": "channel-2",
		"osmosis-ibc-filter-a1": "channel-7",
	}
	for _, chain := range decoded.Chains {
		if chain.CompatMode != "0.37" {
			t.Fatalf("chain %s lost compat_mode: %q", chain.ID, chain.CompatMode)
		}
		if chain.PacketFilter.Policy != "allow" {
			t.Fatalf("chain %s policy = %q, want allow", chain.ID, chain.PacketFilter.Policy)
		}
		want := wantChannels[chain.ID]
		if len(chain.PacketFilter.List) != 1 || len(chain.PacketFilter.List[0]) != 2 || chain.PacketFilter.List[0][0] != "transfer" || chain.PacketFilter.List[0][1] != want {
			t.Fatalf("chain %s filter = %#v, want transfer/%s", chain.ID, chain.PacketFilter.List, want)
		}
	}
}
