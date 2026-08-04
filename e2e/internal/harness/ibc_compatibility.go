package harness

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	IBCCompatibilityMatrixArtifactPath = "ibc-compatibility-matrix.json"
	ibcCompatibilityMatrixSchema       = "panacea.ibc-compatibility-matrix/v1"

	panaceaV221SourceCommit = "a1b342939ba6ac3092aeebbee6a2fa741a34d47f"
	panaceaCurrentVersion   = "2.3.0"

	cosmosSDKModulePath  = "github.com/cosmos/cosmos-sdk"
	cometBFTModulePath   = "github.com/cometbft/cometbft"
	ibcGoV7ModulePath    = "github.com/cosmos/ibc-go/v7"
	ibcGoV8ModulePath    = "github.com/cosmos/ibc-go/v8"
	pfmV8ModulePath      = "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8"
	ibcHooksModulePath   = "github.com/osmosis-labs/osmosis/x/ibc-hooks"
	osmosisSDKModulePath = "github.com/osmosis-labs/cosmos-sdk"

	ibcInvestigationConfirmed               = "confirmed"
	ibcInvestigationPinnedSourceLiveLimited = "pinned-source-live-limited"
	ibcInvestigationMismatch                = "mismatch"
	ibcInvestigationUnavailable             = "unavailable"
	ibcMiddlewareLiveObservationScope       = "channel-and-build-metadata-only"
)

// IBCDependencyContract is an exact Go module path/version pair. Paths are
// intentionally structural: merely finding a version string anywhere in
// `version --long` is insufficient evidence that the expected module supplied
// the executable.
type IBCDependencyContract struct {
	Path        string                    `json:"path"`
	Version     string                    `json:"version"`
	Replacement *IBCDependencyReplacement `json:"replacement,omitempty"`
}

type IBCDependencyReplacement struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

func (d IBCDependencyContract) EffectiveVersion() string {
	if d.Replacement != nil {
		return d.Replacement.Version
	}
	return d.Version
}

// IBCBinaryVersionContract is the immutable identity expected from one chain
// binary. Commit may only be empty for contracts whose upstream command does
// not expose a source commit; all contracts used by this topology require it.
type IBCBinaryVersionContract struct {
	Name             string                  `json:"name"`
	AppName          string                  `json:"app_name"`
	BinaryPath       string                  `json:"binary_path"`
	Version          string                  `json:"version"`
	Commit           string                  `json:"commit"`
	CosmosSDKVersion string                  `json:"cosmos_sdk_version"`
	Dependencies     []IBCDependencyContract `json:"dependencies"`
}

func (c IBCBinaryVersionContract) DependencyVersion(path string) string {
	for _, dependency := range c.Dependencies {
		if dependency.Path == path {
			return dependency.EffectiveVersion()
		}
	}
	return ""
}

func panaceaV221BinaryContract() IBCBinaryVersionContract {
	return IBCBinaryVersionContract{
		Name:             "panacea-core",
		AppName:          "panacead",
		BinaryPath:       "/usr/bin/panacead",
		Version:          "2.2.1",
		Commit:           panaceaV221SourceCommit,
		CosmosSDKVersion: "v0.47.10",
		Dependencies: []IBCDependencyContract{
			{Path: cosmosSDKModulePath, Version: "v0.47.10"},
			{
				Path:    cometBFTModulePath,
				Version: "v0.37.4",
				Replacement: &IBCDependencyReplacement{
					Path: cometBFTModulePath, Version: "v0.37.18",
				},
			},
			{Path: ibcGoV7ModulePath, Version: "v7.3.2"},
		},
	}
}

func currentPanaceaBinaryContract() (IBCBinaryVersionContract, error) {
	version := strings.TrimSpace(os.Getenv("PANACEA_E2E_CURRENT_BINARY_VERSION"))
	if version == "" {
		return IBCBinaryVersionContract{}, errors.New("PANACEA_E2E_CURRENT_BINARY_VERSION is required for IBC compatibility validation")
	}
	if version != panaceaCurrentVersion {
		return IBCBinaryVersionContract{}, fmt.Errorf(
			"PANACEA_E2E_CURRENT_BINARY_VERSION = %q, want exact %q for the IBC upgrade contract",
			version,
			panaceaCurrentVersion,
		)
	}
	commit := strings.TrimSpace(os.Getenv("PANACEA_E2E_CURRENT_COMMIT"))
	if commit == "" {
		return IBCBinaryVersionContract{}, errors.New("PANACEA_E2E_CURRENT_COMMIT is required for IBC compatibility validation")
	}
	if err := validateGitCommit(commit); err != nil {
		return IBCBinaryVersionContract{}, fmt.Errorf("PANACEA_E2E_CURRENT_COMMIT: %w", err)
	}
	if commit == panaceaV221SourceCommit {
		return IBCBinaryVersionContract{}, errors.New("PANACEA_E2E_CURRENT_COMMIT resolves to the v2.2.1 source commit")
	}
	return IBCBinaryVersionContract{
		Name:             "panacea-core",
		AppName:          "panacead",
		BinaryPath:       "/usr/bin/panacead",
		Version:          version,
		Commit:           commit,
		CosmosSDKVersion: "v0.50.15",
		Dependencies: []IBCDependencyContract{
			{Path: cosmosSDKModulePath, Version: "v0.50.15"},
			{Path: cometBFTModulePath, Version: "v0.38.23"},
			{Path: ibcGoV8ModulePath, Version: "v8.8.0"},
		},
	}, nil
}

func pinnedOsmosisBinaryContract() IBCBinaryVersionContract {
	provenance := PinnedIBCProvenance().Osmosis
	return IBCBinaryVersionContract{
		Name:             "osmosis",
		AppName:          "osmosisd",
		BinaryPath:       "/bin/osmosisd",
		Version:          provenance.Tag,
		Commit:           provenance.SourceCommit,
		CosmosSDKVersion: provenance.CosmosSDKVersion,
		Dependencies: []IBCDependencyContract{
			{
				Path:    cosmosSDKModulePath,
				Version: "v0.50.14",
				Replacement: &IBCDependencyReplacement{
					Path: osmosisSDKModulePath, Version: "v0.50.14-v30-osmo",
				},
			},
			{Path: cometBFTModulePath, Version: provenance.CometBFTVersion},
			{Path: ibcGoV8ModulePath, Version: provenance.IBCGoVersion},
			{Path: pfmV8ModulePath, Version: "v8.2.0"},
			{Path: ibcHooksModulePath, Version: "v0.0.21"},
		},
	}
}

// osmosisNodeInfoObservableContract removes replacement metadata because the
// Cosmos node-info REST shape exposes only each original module path/version.
// The exact replacement remains mandatory for local `version --long` and is
// separately pinned in the Osmosis source contract artifact.
func osmosisNodeInfoObservableContract() IBCBinaryVersionContract {
	contract := pinnedOsmosisBinaryContract()
	contract.Dependencies = append([]IBCDependencyContract(nil), contract.Dependencies...)
	for index := range contract.Dependencies {
		contract.Dependencies[index].Replacement = nil
	}
	return contract
}

func panaceaBinaryContractForImage(image ImageRef) (IBCBinaryVersionContract, error) {
	if image == V221Image() {
		return panaceaV221BinaryContract(), nil
	}
	if image == CurrentImage() {
		return currentPanaceaBinaryContract()
	}
	return IBCBinaryVersionContract{}, fmt.Errorf(
		"Panacea IBC image %s:%s has no exact binary compatibility contract",
		image.Repository,
		image.Version,
	)
}

// IBCBinaryVersionIdentity is parsed from the public `version --long` output.
// Dependencies preserve both sides of Go module replacement syntax.
type IBCBinaryVersionIdentity struct {
	Name             string                  `json:"name"`
	AppName          string                  `json:"app_name"`
	Version          string                  `json:"version"`
	Commit           string                  `json:"commit"`
	CosmosSDKVersion string                  `json:"cosmos_sdk_version"`
	GoVersion        string                  `json:"go_version"`
	Dependencies     []IBCDependencyContract `json:"dependencies"`
}

func (i IBCBinaryVersionIdentity) DependencyVersion(path string) string {
	dependency, ok := i.Dependency(path)
	if !ok {
		return ""
	}
	return dependency.EffectiveVersion()
}

func (i IBCBinaryVersionIdentity) Dependency(path string) (IBCDependencyContract, bool) {
	for _, dependency := range i.Dependencies {
		if dependency.Path == path {
			return dependency, true
		}
	}
	return IBCDependencyContract{}, false
}

func parseAndValidateIBCBinaryVersionLong(
	contents []byte,
	contract IBCBinaryVersionContract,
) (IBCBinaryVersionIdentity, error) {
	identity, err := parseIBCBinaryVersionLong(contents)
	if err != nil {
		return IBCBinaryVersionIdentity{}, err
	}
	if err := validateIBCBinaryVersionIdentity(identity, contract); err != nil {
		return identity, err
	}
	return identity, nil
}

func parseIBCBinaryVersionLong(contents []byte) (IBCBinaryVersionIdentity, error) {
	fields := make(map[string]string)
	dependencies := make(map[string]IBCDependencyContract)
	scanner := bufio.NewScanner(strings.NewReader(string(contents)))
	scanner.Buffer(make([]byte, 4096), 4<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- ") {
			dependency, err := parseVersionLongDependency(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			if err != nil {
				return IBCBinaryVersionIdentity{}, err
			}
			if existing, duplicate := dependencies[dependency.Path]; duplicate && !dependencyContractEqual(existing, dependency) {
				return IBCBinaryVersionIdentity{}, fmt.Errorf("version --long contains conflicting dependencies for %s", dependency.Path)
			}
			dependencies[dependency.Path] = dependency
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		switch strings.TrimSpace(key) {
		case "name", "server_name", "version", "commit", "cosmos_sdk_version", "go":
			fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return IBCBinaryVersionIdentity{}, fmt.Errorf("scan version --long: %w", err)
	}

	paths := make([]string, 0, len(dependencies))
	for path := range dependencies {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	parsedDependencies := make([]IBCDependencyContract, 0, len(paths))
	for _, path := range paths {
		parsedDependencies = append(parsedDependencies, dependencies[path])
	}
	identity := IBCBinaryVersionIdentity{
		Name:             fields["name"],
		AppName:          fields["server_name"],
		Version:          fields["version"],
		Commit:           fields["commit"],
		CosmosSDKVersion: fields["cosmos_sdk_version"],
		GoVersion:        fields["go"],
		Dependencies:     parsedDependencies,
	}
	if identity.Name == "" || identity.AppName == "" || identity.Version == "" ||
		identity.Commit == "" || identity.CosmosSDKVersion == "" || identity.GoVersion == "" {
		return identity, fmt.Errorf("incomplete version --long identity: %+v", identity)
	}
	return identity, nil
}

func parseVersionLongDependency(value string) (IBCDependencyContract, error) {
	parts := strings.Split(value, " => ")
	if len(parts) > 2 {
		return IBCDependencyContract{}, fmt.Errorf("malformed version --long dependency %q", value)
	}
	path, version, err := parseVersionLongModuleReference(parts[0])
	if err != nil {
		return IBCDependencyContract{}, err
	}
	dependency := IBCDependencyContract{Path: path, Version: version}
	if len(parts) == 2 {
		replacementPath, replacementVersion, err := parseVersionLongModuleReference(parts[1])
		if err != nil {
			return IBCDependencyContract{}, err
		}
		dependency.Replacement = &IBCDependencyReplacement{Path: replacementPath, Version: replacementVersion}
	}
	return dependency, nil
}

func parseVersionLongModuleReference(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	at := strings.LastIndex(value, "@")
	if at < 1 || at == len(value)-1 {
		return "", "", fmt.Errorf("malformed version --long module reference %q", value)
	}
	path := strings.TrimSpace(value[:at])
	version := strings.Fields(strings.TrimSpace(value[at+1:]))
	if path == "" || len(version) == 0 || version[0] == "" {
		return "", "", fmt.Errorf("malformed version --long module reference %q", value)
	}
	return path, version[0], nil
}

func validateIBCBinaryVersionIdentity(
	identity IBCBinaryVersionIdentity,
	contract IBCBinaryVersionContract,
) error {
	var validationErrors []error
	if identity.Name != contract.Name {
		validationErrors = append(validationErrors, fmt.Errorf("binary name = %q, want %q", identity.Name, contract.Name))
	}
	if identity.AppName != contract.AppName {
		validationErrors = append(validationErrors, fmt.Errorf("binary app name = %q, want %q", identity.AppName, contract.AppName))
	}
	if identity.Version != contract.Version {
		validationErrors = append(validationErrors, fmt.Errorf("binary version = %q, want %q", identity.Version, contract.Version))
	}
	if identity.Commit != contract.Commit {
		validationErrors = append(validationErrors, fmt.Errorf("binary commit = %q, want %q", identity.Commit, contract.Commit))
	}
	if identity.CosmosSDKVersion != contract.CosmosSDKVersion {
		validationErrors = append(validationErrors, fmt.Errorf(
			"binary Cosmos SDK identity = %q, want %q",
			identity.CosmosSDKVersion,
			contract.CosmosSDKVersion,
		))
	}
	for _, expected := range contract.Dependencies {
		observed, ok := identity.Dependency(expected.Path)
		if !ok || !dependencyContractEqual(observed, expected) {
			validationErrors = append(validationErrors, fmt.Errorf(
				"binary dependency %s = %s, want %s",
				expected.Path,
				formatDependencyContract(observed, ok),
				formatDependencyContract(expected, true),
			))
		}
	}
	return errors.Join(validationErrors...)
}

type IBCBinaryNodeEvidence struct {
	Name         string                   `json:"name"`
	Version      IBCBinaryVersionIdentity `json:"version"`
	BinaryPath   string                   `json:"binary_path"`
	BinarySHA256 string                   `json:"binary_sha256"`
}

type IBCChainBinaryEvidence struct {
	ChainID   string                   `json:"chain_id"`
	Phase     string                   `json:"phase"`
	Contract  IBCBinaryVersionContract `json:"contract"`
	Nodes     []IBCBinaryNodeEvidence  `json:"nodes"`
	Validated bool                     `json:"validated"`
	Error     string                   `json:"error,omitempty"`
}

func (e IBCChainBinaryEvidence) Validate() error {
	var validationErrors []error
	if strings.TrimSpace(e.ChainID) == "" || strings.TrimSpace(e.Phase) == "" {
		validationErrors = append(validationErrors, errors.New("chain binary evidence identity is incomplete"))
	}
	if len(e.Nodes) == 0 {
		validationErrors = append(validationErrors, errors.New("chain binary evidence has no nodes"))
	}
	wantChecksum := ""
	seenNodes := make(map[string]struct{}, len(e.Nodes))
	for _, node := range e.Nodes {
		if strings.TrimSpace(node.Name) == "" {
			validationErrors = append(validationErrors, errors.New("chain binary evidence has an unnamed node"))
		}
		if _, duplicate := seenNodes[node.Name]; duplicate {
			validationErrors = append(validationErrors, fmt.Errorf("chain binary evidence duplicates node %q", node.Name))
		}
		seenNodes[node.Name] = struct{}{}
		if err := validateIBCBinaryVersionIdentity(node.Version, e.Contract); err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("node %s: %w", node.Name, err))
		}
		if err := validateSHA256(node.BinarySHA256); err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("node %s binary checksum: %w", node.Name, err))
		}
		if node.BinaryPath != e.Contract.BinaryPath {
			validationErrors = append(validationErrors, fmt.Errorf("node %s binary path = %q, want exact %q", node.Name, node.BinaryPath, e.Contract.BinaryPath))
		}
		if wantChecksum == "" {
			wantChecksum = node.BinarySHA256
		} else if node.BinarySHA256 != wantChecksum {
			validationErrors = append(validationErrors, errors.New("nodes running one chain image have different binary checksums"))
		}
	}
	if !e.Validated {
		validationErrors = append(validationErrors, errors.New("chain binary evidence is not marked validated"))
	}
	if e.Error != "" {
		validationErrors = append(validationErrors, fmt.Errorf("chain binary evidence recorded an error: %s", e.Error))
	}
	return errors.Join(validationErrors...)
}

func parseBinaryChecksumOutput(output []byte, appName string) (string, string, error) {
	fields := strings.Fields(string(output))
	if len(fields) != 2 {
		return "", "", errors.New("binary checksum output must contain one SHA-256 and one path")
	}
	if err := validateSHA256(fields[0]); err != nil {
		return "", "", err
	}
	if !filepath.IsAbs(fields[1]) || filepath.Base(fields[1]) != appName {
		return "", "", fmt.Errorf("binary checksum path %q is not an absolute %s path", fields[1], appName)
	}
	return fields[0], fields[1], nil
}

func validateSHA256(value string) error {
	if value != strings.ToLower(value) {
		return errors.New("SHA-256 must be lowercase")
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		return errors.New("value is not a lowercase SHA-256")
	}
	return nil
}

func validateGitCommit(value string) error {
	if value != strings.ToLower(value) {
		return errors.New("Git commit must be lowercase")
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 20 {
		return errors.New("Git commit must be a full 40-character hexadecimal object ID")
	}
	return nil
}

type IBCGenesisNodeChecksum struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type IBCGenesisChecksumSnapshot struct {
	Phase  string                   `json:"phase"`
	Nodes  []IBCGenesisNodeChecksum `json:"nodes"`
	Common string                   `json:"common_sha256"`
}

func (s IBCGenesisChecksumSnapshot) Validate() error {
	var validationErrors []error
	if strings.TrimSpace(s.Phase) == "" || len(s.Nodes) == 0 {
		validationErrors = append(validationErrors, errors.New("genesis checksum snapshot is incomplete"))
	}
	if err := validateSHA256(s.Common); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("common genesis checksum: %w", err))
	}
	seenNodes := make(map[string]struct{}, len(s.Nodes))
	for _, node := range s.Nodes {
		if _, duplicate := seenNodes[node.Name]; duplicate {
			validationErrors = append(validationErrors, fmt.Errorf("genesis checksum snapshot duplicates node %q", node.Name))
		}
		seenNodes[node.Name] = struct{}{}
		if strings.TrimSpace(node.Name) == "" || node.SHA256 != s.Common {
			validationErrors = append(validationErrors, fmt.Errorf("node %q genesis checksum %q does not match common checksum %q", node.Name, node.SHA256, s.Common))
		}
	}
	return errors.Join(validationErrors...)
}

type IBCGenesisImmutabilityEvidence struct {
	ChainID     string                     `json:"chain_id"`
	File        string                     `json:"file"`
	Initial     IBCGenesisChecksumSnapshot `json:"initial"`
	PostUpgrade IBCGenesisChecksumSnapshot `json:"post_upgrade"`
	Immutable   bool                       `json:"immutable"`
}

func (e IBCGenesisImmutabilityEvidence) Validate() error {
	var validationErrors []error
	if strings.TrimSpace(e.ChainID) == "" || e.File != "config/genesis.json" {
		validationErrors = append(validationErrors, errors.New("Osmosis genesis identity is incomplete"))
	}
	if err := e.Initial.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("initial genesis: %w", err))
	}
	if err := e.PostUpgrade.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("post-upgrade genesis: %w", err))
	}
	if e.Initial.Common != e.PostUpgrade.Common {
		validationErrors = append(validationErrors, errors.New("Osmosis genesis checksum changed during the Panacea upgrade"))
	}
	if !genesisNodeSetsEqual(e.Initial, e.PostUpgrade) {
		validationErrors = append(validationErrors, errors.New("Osmosis genesis checksum phases do not cover the same nodes"))
	}
	if !e.Immutable {
		validationErrors = append(validationErrors, errors.New("Osmosis genesis evidence is not marked immutable"))
	}
	return errors.Join(validationErrors...)
}

func checksumGenesis(name string, contents []byte) IBCGenesisNodeChecksum {
	digest := sha256.Sum256(contents)
	return IBCGenesisNodeChecksum{Name: name, SHA256: hex.EncodeToString(digest[:])}
}

// IBCMiddlewareInvestigationEvidence deliberately distinguishes what the live
// APIs expose from what is known only from the exact pinned source tree. Live
// node-info exposes compiled dependency versions and the channel endpoint
// exposes the negotiated app version; neither proves runtime middleware wiring
// or per-channel activation.
type IBCMiddlewareInvestigationEvidence struct {
	InvestigationStatus              string                  `json:"investigation_status"`
	LiveObservationScope             string                  `json:"live_observation_scope,omitempty"`
	ChannelApplicationVersion        string                  `json:"channel_application_version,omitempty"`
	WiringObservedLive               bool                    `json:"wiring_observed_live"`
	PerChannelActivationObservedLive bool                    `json:"per_channel_activation_observed_live"`
	SourceContractArtifact           string                  `json:"source_contract_artifact,omitempty"`
	SourceCommit                     string                  `json:"source_commit,omitempty"`
	SourceReference                  string                  `json:"source_reference,omitempty"`
	RecvStack                        []string                `json:"recv_stack,omitempty"`
	SendStack                        []string                `json:"send_stack,omitempty"`
	Dependencies                     []IBCDependencyContract `json:"dependencies,omitempty"`
	Error                            string                  `json:"error,omitempty"`
}

func expectedOsmosisMiddlewareEvidence() IBCMiddlewareInvestigationEvidence {
	source := pinnedOsmosisSourceContract()
	return IBCMiddlewareInvestigationEvidence{
		InvestigationStatus:              ibcInvestigationPinnedSourceLiveLimited,
		LiveObservationScope:             ibcMiddlewareLiveObservationScope,
		ChannelApplicationVersion:        "ics20-1",
		WiringObservedLive:               false,
		PerChannelActivationObservedLive: false,
		SourceContractArtifact:           OsmosisPinnedSourceContractArtifactPath,
		SourceCommit:                     source.Commit,
		SourceReference:                  source.TransferWiring.Reference,
		RecvStack:                        append([]string(nil), source.RecvStack...),
		SendStack:                        append([]string(nil), source.SendStack...),
		Dependencies: []IBCDependencyContract{
			{Path: pfmV8ModulePath, Version: "v8.2.0"},
			{Path: ibcHooksModulePath, Version: "v0.0.21"},
		},
	}
}

func (e IBCMiddlewareInvestigationEvidence) Validate() error {
	want := expectedOsmosisMiddlewareEvidence()
	var validationErrors []error
	if e.InvestigationStatus != ibcInvestigationPinnedSourceLiveLimited {
		validationErrors = append(validationErrors, fmt.Errorf(
			"middleware investigation status = %q, want %q",
			e.InvestigationStatus,
			ibcInvestigationPinnedSourceLiveLimited,
		))
	}
	if e.LiveObservationScope != ibcMiddlewareLiveObservationScope || e.ChannelApplicationVersion != "ics20-1" {
		validationErrors = append(validationErrors, errors.New("middleware live evidence is not limited to channel and build metadata"))
	}
	if e.WiringObservedLive || e.PerChannelActivationObservedLive {
		validationErrors = append(validationErrors, errors.New("middleware live evidence overstates wiring or per-channel activation visibility"))
	}
	if e.SourceContractArtifact != OsmosisPinnedSourceContractArtifactPath ||
		e.SourceCommit != want.SourceCommit || e.SourceReference != want.SourceReference {
		validationErrors = append(validationErrors, errors.New("middleware source wiring is not pinned to the Osmosis release commit"))
	}
	if !stringSlicesEqual(e.RecvStack, want.RecvStack) || !stringSlicesEqual(e.SendStack, want.SendStack) {
		validationErrors = append(validationErrors, errors.New("Osmosis ICS-20 middleware order does not match the pinned source wiring"))
	}
	if !dependencyContractsEqual(e.Dependencies, want.Dependencies) {
		validationErrors = append(validationErrors, errors.New("Osmosis middleware dependency versions do not match the pinned binary contract"))
	}
	if e.Error != "" {
		validationErrors = append(validationErrors, fmt.Errorf("middleware investigation recorded an error: %s", e.Error))
	}
	return errors.Join(validationErrors...)
}

func newOsmosisMiddlewareEvidence(
	nodeInfo OsmosisMainnetNodeInfoEvidence,
	channel OsmosisMainnetChannelEvidence,
) (IBCMiddlewareInvestigationEvidence, error) {
	evidence := expectedOsmosisMiddlewareEvidence()
	evidence.ChannelApplicationVersion = channel.Version
	evidence.Dependencies = nil
	for _, path := range []string{pfmV8ModulePath, ibcHooksModulePath} {
		dependency, ok := nodeInfo.Binary.Dependency(path)
		if !ok {
			return evidence, fmt.Errorf("live Osmosis node-info has no middleware dependency %s", path)
		}
		evidence.Dependencies = append(evidence.Dependencies, dependency)
	}
	if err := evidence.Validate(); err != nil {
		return evidence, err
	}
	return evidence, nil
}

type IBCCompatibilityChannelContract struct {
	PanaceaChannelID string `json:"panacea_channel_id"`
	OsmosisChannelID string `json:"osmosis_channel_id"`
	PortID           string `json:"port_id"`
	State            string `json:"state"`
	Ordering         string `json:"ordering"`
	Version          string `json:"version"`
}

func (c IBCCompatibilityChannelContract) Validate() error {
	var validationErrors []error
	if strings.TrimSpace(c.PanaceaChannelID) == "" || strings.TrimSpace(c.OsmosisChannelID) == "" {
		validationErrors = append(validationErrors, errors.New("local IBC channel IDs are required"))
	}
	if c.PortID != "transfer" || !ibcStateEquals(c.State, "OPEN") ||
		!ibcStateEquals(c.Ordering, "UNORDERED") || c.Version != "ics20-1" {
		validationErrors = append(validationErrors, fmt.Errorf(
			"local channel = %s/%s/%s/%s, want transfer/open/unordered/ics20-1",
			c.PortID,
			c.State,
			c.Ordering,
			c.Version,
		))
	}
	return errors.Join(validationErrors...)
}

type IBCBinaryUpgradeMatrix struct {
	PreUpgrade  IBCChainBinaryEvidence `json:"pre_upgrade"`
	PostUpgrade IBCChainBinaryEvidence `json:"post_upgrade"`
}

type IBCOsmosisCompatibilityMatrix struct {
	SourceContract OsmosisPinnedSourceContract    `json:"source_contract"`
	PreUpgrade     IBCChainBinaryEvidence         `json:"pre_upgrade"`
	PostUpgrade    IBCChainBinaryEvidence         `json:"post_upgrade"`
	Genesis        IBCGenesisImmutabilityEvidence `json:"genesis"`
}

type IBCHermesCompatibilityMatrix struct {
	RuntimeIdentity HermesRuntimeIdentityEvidence `json:"runtime_identity"`
	CompatMode      map[string]string             `json:"compat_mode"`
}

type IBCCompatibilityMatrix struct {
	SchemaVersion    string                             `json:"schema_version"`
	GeneratedAt      time.Time                          `json:"generated_at"`
	MainnetPreflight OsmosisMainnetPreflightEvidence    `json:"mainnet_preflight"`
	Panacea          IBCBinaryUpgradeMatrix             `json:"panacea"`
	Osmosis          IBCOsmosisCompatibilityMatrix      `json:"osmosis"`
	Channel          IBCCompatibilityChannelContract    `json:"channel"`
	Middleware       IBCMiddlewareInvestigationEvidence `json:"middleware"`
	Hermes           IBCHermesCompatibilityMatrix       `json:"hermes"`
	Validated        bool                               `json:"validated"`
	Error            string                             `json:"error,omitempty"`
}

func (m IBCCompatibilityMatrix) Validate() error {
	var validationErrors []error
	if m.SchemaVersion != ibcCompatibilityMatrixSchema || m.GeneratedAt.IsZero() {
		validationErrors = append(validationErrors, errors.New("IBC compatibility matrix metadata is incomplete"))
	}
	if err := m.MainnetPreflight.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("mainnet preflight: %w", err))
	}
	if err := m.Panacea.PreUpgrade.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Panacea pre-upgrade binary: %w", err))
	}
	if err := m.Panacea.PostUpgrade.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Panacea post-upgrade binary: %w", err))
	}
	if m.Panacea.PreUpgrade.Phase != "pre-upgrade" || m.Panacea.PostUpgrade.Phase != "post-upgrade" ||
		m.Panacea.PreUpgrade.ChainID != m.Panacea.PostUpgrade.ChainID {
		validationErrors = append(validationErrors, errors.New("Panacea binary phases do not describe one pre/post-upgrade chain"))
	}
	if !binaryContractsEqual(m.Panacea.PreUpgrade.Contract, panaceaV221BinaryContract()) {
		validationErrors = append(validationErrors, errors.New("Panacea pre-upgrade dependency contract is not v2.2.1"))
	}
	currentContract, err := currentPanaceaBinaryContract()
	if err != nil {
		validationErrors = append(validationErrors, err)
	} else if !binaryContractsEqual(m.Panacea.PostUpgrade.Contract, currentContract) {
		validationErrors = append(validationErrors, errors.New("Panacea post-upgrade dependency contract is not the requested current build"))
	}
	if chainBinaryChecksumsEqual(m.Panacea.PreUpgrade, m.Panacea.PostUpgrade) {
		validationErrors = append(validationErrors, errors.New("Panacea binary checksum did not change across the binary upgrade"))
	}
	if err := m.Osmosis.PreUpgrade.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis pre-upgrade binary: %w", err))
	}
	if err := m.Osmosis.PostUpgrade.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis post-upgrade binary: %w", err))
	}
	if m.Osmosis.PreUpgrade.Phase != "pre-upgrade" || m.Osmosis.PostUpgrade.Phase != "post-upgrade" ||
		m.Osmosis.PreUpgrade.ChainID != m.Osmosis.PostUpgrade.ChainID ||
		m.Osmosis.Genesis.ChainID != m.Osmosis.PreUpgrade.ChainID {
		validationErrors = append(validationErrors, errors.New("Osmosis binary/genesis phases do not describe one pre/post-upgrade chain"))
	}
	if err := m.Osmosis.SourceContract.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis pinned source contract: %w", err))
	}
	if m.Osmosis.SourceContract.Commit != m.MainnetPreflight.NodeInfo.Binary.Commit ||
		m.Osmosis.SourceContract.Commit != m.Osmosis.PreUpgrade.Contract.Commit {
		validationErrors = append(validationErrors, errors.New("Osmosis source, live-mainnet, and local binary commits do not match"))
	}
	if m.Panacea.PreUpgrade.ChainID == m.Osmosis.PreUpgrade.ChainID {
		validationErrors = append(validationErrors, errors.New("Panacea and Osmosis compatibility contracts must use distinct chains"))
	}
	if !binaryContractsEqual(m.Osmosis.PreUpgrade.Contract, pinnedOsmosisBinaryContract()) ||
		!binaryContractsEqual(m.Osmosis.PostUpgrade.Contract, pinnedOsmosisBinaryContract()) {
		validationErrors = append(validationErrors, errors.New("Osmosis binary contracts are not the pinned v31.0.2 release"))
	}
	if !chainBinaryChecksumsEqual(m.Osmosis.PreUpgrade, m.Osmosis.PostUpgrade) {
		validationErrors = append(validationErrors, errors.New("Osmosis binary checksums changed during the Panacea upgrade"))
	}
	if err := m.Osmosis.Genesis.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis genesis: %w", err))
	}
	if err := m.Channel.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("channel: %w", err))
	}
	if err := m.Middleware.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("middleware: %w", err))
	}
	validatedHermes, hermesErr := validateHermesRuntimeIdentity(
		m.Hermes.RuntimeIdentity.VersionOutput,
		m.Hermes.RuntimeIdentity.BinarySHA256+"  "+m.Hermes.RuntimeIdentity.BinaryPath,
		PinnedIBCProvenance(),
	)
	if hermesErr != nil || validatedHermes != m.Hermes.RuntimeIdentity {
		validationErrors = append(validationErrors, errors.New("Hermes runtime identity is not the pinned release and binary checksum"))
	}
	if len(m.Hermes.CompatMode) != 2 {
		validationErrors = append(validationErrors, errors.New("Hermes compat_mode must cover exactly both local chains"))
	}
	for _, chainID := range []string{m.Panacea.PreUpgrade.ChainID, m.Osmosis.PreUpgrade.ChainID} {
		if mode := m.Hermes.CompatMode[chainID]; mode != "0.37" {
			validationErrors = append(validationErrors, fmt.Errorf("Hermes compat_mode[%q] = %q, want 0.37", chainID, mode))
		}
	}
	if !m.Validated {
		validationErrors = append(validationErrors, errors.New("IBC compatibility matrix is not marked validated"))
	}
	if m.Error != "" {
		validationErrors = append(validationErrors, fmt.Errorf("IBC compatibility matrix recorded an error: %s", m.Error))
	}
	return errors.Join(validationErrors...)
}

func binaryContractsEqual(left, right IBCBinaryVersionContract) bool {
	return left.Name == right.Name && left.AppName == right.AppName && left.BinaryPath == right.BinaryPath && left.Version == right.Version &&
		left.Commit == right.Commit && left.CosmosSDKVersion == right.CosmosSDKVersion &&
		dependencyContractsEqual(left.Dependencies, right.Dependencies)
}

func dependencyContractsEqual(left, right []IBCDependencyContract) bool {
	if len(left) != len(right) {
		return false
	}
	leftMap := make(map[string]IBCDependencyContract, len(left))
	for _, dependency := range left {
		if _, duplicate := leftMap[dependency.Path]; duplicate {
			return false
		}
		leftMap[dependency.Path] = dependency
	}
	for _, dependency := range right {
		observed, ok := leftMap[dependency.Path]
		if !ok || !dependencyContractEqual(observed, dependency) {
			return false
		}
	}
	return true
}

func dependencyContractEqual(left, right IBCDependencyContract) bool {
	if left.Path != right.Path || left.Version != right.Version {
		return false
	}
	if left.Replacement == nil || right.Replacement == nil {
		return left.Replacement == nil && right.Replacement == nil
	}
	return *left.Replacement == *right.Replacement
}

func formatDependencyContract(dependency IBCDependencyContract, available bool) string {
	if !available {
		return "<missing>"
	}
	formatted := dependency.Path + "@" + dependency.Version
	if dependency.Replacement != nil {
		formatted += " => " + dependency.Replacement.Path + "@" + dependency.Replacement.Version
	}
	return formatted
}

func stringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func chainBinaryChecksumsEqual(left, right IBCChainBinaryEvidence) bool {
	if len(left.Nodes) == 0 || len(left.Nodes) != len(right.Nodes) {
		return false
	}
	leftChecksums := make(map[string]string, len(left.Nodes))
	for _, node := range left.Nodes {
		leftChecksums[node.Name] = node.BinarySHA256
	}
	for _, node := range right.Nodes {
		if leftChecksums[node.Name] != node.BinarySHA256 {
			return false
		}
	}
	return true
}

func genesisNodeSetsEqual(left, right IBCGenesisChecksumSnapshot) bool {
	if len(left.Nodes) == 0 || len(left.Nodes) != len(right.Nodes) {
		return false
	}
	leftNodes := make(map[string]struct{}, len(left.Nodes))
	for _, node := range left.Nodes {
		leftNodes[node.Name] = struct{}{}
	}
	for _, node := range right.Nodes {
		if _, ok := leftNodes[node.Name]; !ok {
			return false
		}
	}
	return true
}
