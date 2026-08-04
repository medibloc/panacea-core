package harness

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	P0P1ReleaseGateSchemaVersion       = "2"
	P0P1ReleaseGateArtifactPath        = "release/gate-manifest.json"
	P0P1ReleaseGateFailureArtifactPath = "release/gate-failure.json"
)

var requiredP0P1ReleaseGateTests = []string{
	"TestV221ToCurrentMultiValidatorUpgrade",
	"TestV221ToCurrentLegacyPNFTAdversarialUpgrade",
	"TestV221UpgradeBoundaryChaos/seed-101",
	"TestV221UpgradeBoundaryChaos/seed-202",
	"TestV221UpgradeBoundaryChaos/seed-303",
	"TestIBCUpgradeContinuity",
	"TestActualCometStateSyncAndBadTrustHash",
	"TestV047NodeHomeConfigCompatibility",
	"TestLocalDockerNetworkAndEndpointFaults",
}

var oldImageP0P1ReleaseGateTests = map[string]struct{}{
	"TestV221ToCurrentMultiValidatorUpgrade":        {},
	"TestV221ToCurrentLegacyPNFTAdversarialUpgrade": {},
	"TestV221UpgradeBoundaryChaos/seed-101":         {},
	"TestV221UpgradeBoundaryChaos/seed-202":         {},
	"TestV221UpgradeBoundaryChaos/seed-303":         {},
	"TestIBCUpgradeContinuity":                      {},
	"TestV047NodeHomeConfigCompatibility":           {},
}

// P0P1ReleaseGateSuiteEvidence ties one required live suite to its successful,
// cleaned run and to the content-addressed Docker image IDs it actually used.
type P0P1ReleaseGateSuiteEvidence struct {
	TestName        string   `json:"test_name"`
	RunPath         string   `json:"run_path"`
	InitialImageRef string   `json:"initial_image_ref"`
	FinalImageIDs   []string `json:"final_image_ids"`
	OldImageIDs     []string `json:"old_image_ids,omitempty"`
	NodeCount       int      `json:"node_count"`
	SwitchCount     int      `json:"switch_count"`
	SwitchedNodes   []string `json:"switched_nodes,omitempty"`
}

// P0P1ReleaseGateManifest is written only after every P0/P1 live prerequisite,
// the 13-row coverage merge, and both release-architecture upgrades succeed.
// Docker image IDs are sha256 content addresses, not mutable repository tags.
type P0P1ReleaseGateManifest struct {
	SchemaVersion          string                          `json:"schema_version"`
	RecordedAt             time.Time                       `json:"recorded_at"`
	SourceCommit           string                          `json:"source_commit"`
	SourceClean            bool                            `json:"source_clean"`
	CoverageMatrix         string                          `json:"coverage_matrix"`
	ReleaseBuildManifest   string                          `json:"release_build_manifest"`
	ReleaseHostPlatform    string                          `json:"release_host_platform"`
	ReleaseHostIdentity    string                          `json:"release_host_identity"`
	ReleaseHostImages      []ReleaseHostImageIdentityEntry `json:"release_host_images"`
	ReleaseImages          []ReleasePlatformImageEvidence  `json:"release_images"`
	CurrentImageID         string                          `json:"current_image_id"`
	OldImageID             string                          `json:"old_image_id"`
	CurrentInitialImageRef string                          `json:"current_initial_image_ref"`
	OldInitialImageRef     string                          `json:"old_initial_image_ref"`
	RequiredSuites         []P0P1ReleaseGateSuiteEvidence  `json:"required_suites"`
}

type P0P1ReleaseGateFailure struct {
	SchemaVersion string    `json:"schema_version"`
	RecordedAt    time.Time `json:"recorded_at"`
	Stage         string    `json:"stage"`
	Error         string    `json:"error"`
}

func (m P0P1ReleaseGateManifest) Validate() error {
	if m.SchemaVersion != P0P1ReleaseGateSchemaVersion {
		return fmt.Errorf("P0/P1 release gate schema version %q, want %q", m.SchemaVersion, P0P1ReleaseGateSchemaVersion)
	}
	if m.RecordedAt.IsZero() {
		return errors.New("P0/P1 release gate recorded_at is required")
	}
	if !releaseCommitPattern.MatchString(m.SourceCommit) {
		return fmt.Errorf("P0/P1 release gate source commit %q is invalid", m.SourceCommit)
	}
	if !m.SourceClean {
		return errors.New("P0/P1 release gate requires clean HEAD source evidence")
	}
	if m.CoverageMatrix != UpgradeCoverageMatrixArtifactPath {
		return fmt.Errorf("P0/P1 release gate coverage matrix %q, want %q", m.CoverageMatrix, UpgradeCoverageMatrixArtifactPath)
	}
	if strings.TrimSpace(m.ReleaseBuildManifest) == "" {
		return errors.New("P0/P1 release build manifest path is required")
	}
	if err := validateCoverageArtifactPath(m.ReleaseBuildManifest); err != nil {
		return fmt.Errorf("P0/P1 release build manifest: %w", err)
	}
	if !containsReleasePlatform(m.ReleaseHostPlatform) {
		return fmt.Errorf("P0/P1 release host platform %q is invalid", m.ReleaseHostPlatform)
	}
	expectedHostIdentity := filepath.ToSlash(filepath.Join(filepath.Dir(m.ReleaseBuildManifest), ReleaseHostImageIdentityArtifactPath))
	if m.ReleaseHostIdentity != expectedHostIdentity {
		return fmt.Errorf("P0/P1 release host identity %q, want %q", m.ReleaseHostIdentity, expectedHostIdentity)
	}
	releaseIdentity := ReleaseHostImageIdentity{
		SchemaVersion: ReleaseHostImageIdentitySchemaVersion,
		HostPlatform:  m.ReleaseHostPlatform,
		Images:        m.ReleaseHostImages,
	}
	if err := releaseIdentity.Validate(); err != nil {
		return fmt.Errorf("P0/P1 release host identity: %w", err)
	}
	releaseRunID := filepath.Base(filepath.Dir(filepath.FromSlash(m.ReleaseBuildManifest)))
	if err := validateReleasePlatformImages(m.ReleaseImages, releaseRunID, m.SourceCommit); err != nil {
		return fmt.Errorf("P0/P1 release platform images: %w", err)
	}
	if !releaseDigestPattern.MatchString(m.CurrentImageID) || !releaseDigestPattern.MatchString(m.OldImageID) {
		return errors.New("P0/P1 release gate requires sha256 current and old Docker image IDs")
	}
	if m.CurrentImageID == m.OldImageID {
		return errors.New("P0/P1 current and old Docker image IDs must differ")
	}
	if strings.TrimSpace(m.CurrentInitialImageRef) == "" || strings.TrimSpace(m.OldInitialImageRef) == "" {
		return errors.New("P0/P1 initial image references are required")
	}
	if len(m.RequiredSuites) != len(requiredP0P1ReleaseGateTests) {
		return fmt.Errorf("P0/P1 release gate has %d suites, want %d", len(m.RequiredSuites), len(requiredP0P1ReleaseGateTests))
	}
	seen := make(map[string]struct{}, len(m.RequiredSuites))
	for _, suite := range m.RequiredSuites {
		if _, duplicate := seen[suite.TestName]; duplicate {
			return fmt.Errorf("P0/P1 release gate contains duplicate suite %q", suite.TestName)
		}
		seen[suite.TestName] = struct{}{}
		if strings.TrimSpace(suite.RunPath) == "" || filepath.IsAbs(suite.RunPath) || strings.Contains(suite.RunPath, "..") {
			return fmt.Errorf("P0/P1 suite %q has unsafe run path %q", suite.TestName, suite.RunPath)
		}
		if len(suite.FinalImageIDs) != 1 || suite.FinalImageIDs[0] != m.CurrentImageID {
			return fmt.Errorf("P0/P1 suite %q final image IDs %v do not identify current image %s", suite.TestName, suite.FinalImageIDs, m.CurrentImageID)
		}
		if suite.NodeCount < 1 {
			return fmt.Errorf("P0/P1 suite %q node count is %d, want at least one", suite.TestName, suite.NodeCount)
		}
		_, startsOld := oldImageP0P1ReleaseGateTests[suite.TestName]
		if startsOld {
			if suite.InitialImageRef != m.OldInitialImageRef || len(suite.OldImageIDs) != 1 || suite.OldImageIDs[0] != m.OldImageID || suite.SwitchCount != suite.NodeCount || len(suite.SwitchedNodes) != suite.NodeCount {
				return fmt.Errorf("P0/P1 upgrade suite %q does not prove the shared old-to-current image switch", suite.TestName)
			}
			if err := validateP0P1SwitchedNodes(suite.SwitchedNodes); err != nil {
				return fmt.Errorf("P0/P1 upgrade suite %q switched nodes: %w", suite.TestName, err)
			}
		} else if suite.InitialImageRef != m.CurrentInitialImageRef || len(suite.OldImageIDs) != 0 || suite.SwitchCount != 0 || len(suite.SwitchedNodes) != 0 {
			return fmt.Errorf("P0/P1 current-only suite %q has unexpected upgrade image evidence", suite.TestName)
		}
	}
	for _, required := range requiredP0P1ReleaseGateTests {
		if _, ok := seen[required]; !ok {
			return fmt.Errorf("P0/P1 release gate is missing suite %q", required)
		}
	}
	currentReleaseImage, ok := releaseIdentity.image("current")
	if !ok || currentReleaseImage.FunctionalImageID != m.CurrentImageID || currentReleaseImage.FunctionalImageRef != m.CurrentInitialImageRef {
		return errors.New("P0/P1 current functional suite image does not match release host identity")
	}
	oldReleaseImage, ok := releaseIdentity.image("v2.2.1")
	if !ok || oldReleaseImage.FunctionalImageID != m.OldImageID || oldReleaseImage.FunctionalImageRef != m.OldInitialImageRef {
		return errors.New("P0/P1 old functional suite image does not match release host identity")
	}
	for _, hostImage := range releaseIdentity.Images {
		var platformImage *ReleasePlatformImageEvidence
		for index := range m.ReleaseImages {
			candidate := &m.ReleaseImages[index]
			if candidate.Kind == hostImage.Kind && candidate.Platform == m.ReleaseHostPlatform {
				platformImage = candidate
				break
			}
		}
		if platformImage == nil || hostImage.ReleaseImageRef != platformImage.ImageRef ||
			hostImage.ReleaseImageID != platformImage.ImageID ||
			hostImage.ReleaseBinarySHA256 != platformImage.BinarySHA256 {
			return fmt.Errorf("P0/P1 host release image %q does not match platform digest evidence", hostImage.Kind)
		}
	}
	return nil
}

func validateP0P1SwitchedNodes(nodes []string) error {
	seen := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node) == "" {
			return errors.New("node name is empty")
		}
		if _, duplicate := seen[node]; duplicate {
			return fmt.Errorf("node %q appears more than once", node)
		}
		seen[node] = struct{}{}
	}
	sorted := append([]string(nil), nodes...)
	sort.Strings(sorted)
	if strings.Join(sorted, "\x00") != strings.Join(nodes, "\x00") {
		return errors.New("node names are not sorted")
	}
	return nil
}

type p0p1RunManifest struct {
	TestName      string `json:"test_name"`
	Image         string `json:"image"`
	NumValidators int    `json:"num_validators"`
	NumFullNodes  int    `json:"num_full_nodes"`
}

type p0p1ContainerState struct {
	Image string `json:"image"`
}

type p0p1SwitchRecord struct {
	Plan struct {
		Node string `json:"node"`
	} `json:"plan"`
	OldImageID string `json:"old_image_id"`
	NewImageID string `json:"new_image_id"`
	Error      string `json:"error,omitempty"`
}

// WriteP0P1ReleaseGateManifest verifies the complete aggregate root and writes
// its immutable evidence index atomically. It never accepts evidence from a
// sibling root or from a failed/unclean run.
func WriteP0P1ReleaseGateManifest(root, sourceCommit string, recordedAt time.Time) (P0P1ReleaseGateManifest, error) {
	var zero P0P1ReleaseGateManifest
	if !releaseCommitPattern.MatchString(sourceCommit) {
		return zero, fmt.Errorf("P0/P1 source commit %q is invalid", sourceCommit)
	}
	if recordedAt.IsZero() {
		return zero, errors.New("P0/P1 release gate recorded_at is required")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return zero, err
	}
	if err := validateMergedCoverageForReleaseGate(root); err != nil {
		return zero, err
	}
	releaseManifest, releaseRelative, err := findP0P1ReleaseBuildManifest(root, sourceCommit)
	if err != nil {
		return zero, err
	}
	releaseDir := filepath.Dir(filepath.Join(root, filepath.FromSlash(releaseRelative)))
	releaseIdentity, err := readReleaseHostImageIdentity(releaseDir, releaseManifest)
	if err != nil {
		return zero, fmt.Errorf("read P0/P1 release host image identity: %w", err)
	}

	suites, err := discoverP0P1ReleaseGateSuites(root)
	if err != nil {
		return zero, err
	}
	currentIDs := make(map[string]struct{})
	oldIDs := make(map[string]struct{})
	currentRefs := make(map[string]struct{})
	oldRefs := make(map[string]struct{})
	for _, suite := range suites {
		for _, imageID := range suite.FinalImageIDs {
			currentIDs[imageID] = struct{}{}
		}
		if _, startsOld := oldImageP0P1ReleaseGateTests[suite.TestName]; startsOld {
			oldRefs[suite.InitialImageRef] = struct{}{}
			for _, imageID := range suite.OldImageIDs {
				oldIDs[imageID] = struct{}{}
			}
		} else {
			currentRefs[suite.InitialImageRef] = struct{}{}
		}
	}
	currentID, err := requireOneP0P1Value(currentIDs, "current Docker image ID")
	if err != nil {
		return zero, err
	}
	oldID, err := requireOneP0P1Value(oldIDs, "old Docker image ID")
	if err != nil {
		return zero, err
	}
	currentRef, err := requireOneP0P1Value(currentRefs, "current initial image reference")
	if err != nil {
		return zero, err
	}
	oldRef, err := requireOneP0P1Value(oldRefs, "old initial image reference")
	if err != nil {
		return zero, err
	}
	manifest := P0P1ReleaseGateManifest{
		SchemaVersion:          P0P1ReleaseGateSchemaVersion,
		RecordedAt:             recordedAt.UTC(),
		SourceCommit:           sourceCommit,
		SourceClean:            releaseManifest.SourceClean,
		CoverageMatrix:         UpgradeCoverageMatrixArtifactPath,
		ReleaseBuildManifest:   releaseRelative,
		ReleaseHostPlatform:    releaseIdentity.HostPlatform,
		ReleaseHostIdentity:    filepath.ToSlash(filepath.Join(filepath.Dir(releaseRelative), releaseManifest.HostImageIdentity)),
		ReleaseHostImages:      append([]ReleaseHostImageIdentityEntry(nil), releaseIdentity.Images...),
		ReleaseImages:          append([]ReleasePlatformImageEvidence(nil), releaseManifest.Images...),
		CurrentImageID:         currentID,
		OldImageID:             oldID,
		CurrentInitialImageRef: currentRef,
		OldInitialImageRef:     oldRef,
		RequiredSuites:         suites,
	}
	if err := manifest.Validate(); err != nil {
		return zero, err
	}
	if err := writeP0P1ReleaseGateJSON(root, P0P1ReleaseGateArtifactPath, manifest); err != nil {
		return zero, err
	}
	return manifest, nil
}

// WriteP0P1ReleaseGateFailure preserves aggregate-level failures which happen
// after the individual run manifests have already been cleaned and written.
func WriteP0P1ReleaseGateFailure(root, stage string, cause error, recordedAt time.Time) error {
	if strings.TrimSpace(stage) == "" {
		return errors.New("P0/P1 release gate failure stage is required")
	}
	if cause == nil {
		return errors.New("P0/P1 release gate failure cause is required")
	}
	if recordedAt.IsZero() {
		return errors.New("P0/P1 release gate failure recorded_at is required")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	return writeP0P1ReleaseGateJSON(root, P0P1ReleaseGateFailureArtifactPath, P0P1ReleaseGateFailure{
		SchemaVersion: P0P1ReleaseGateSchemaVersion,
		RecordedAt:    recordedAt.UTC(),
		Stage:         strings.TrimSpace(stage),
		Error:         cause.Error(),
	})
}

func validateMergedCoverageForReleaseGate(root string) error {
	var matrix UpgradeCoverageMatrix
	if err := readUpgradeCoverageRunJSON(filepath.Join(root, filepath.FromSlash(UpgradeCoverageMatrixArtifactPath)), &matrix); err != nil {
		return fmt.Errorf("read merged P0/P1 coverage matrix: %w", err)
	}
	if err := matrix.Validate(); err != nil {
		return fmt.Errorf("validate merged P0/P1 coverage matrix: %w", err)
	}
	if len(matrix.SourceMatrices) == 0 {
		return errors.New("P0/P1 release gate requires an aggregate coverage matrix")
	}
	for _, row := range matrix.Rows {
		if row.Status != UpgradeCoverageStatusPassed {
			return fmt.Errorf("P0/P1 coverage row %q is %q, want passed", row.Area, row.Status)
		}
	}
	return nil
}

func findP0P1ReleaseBuildManifest(root, sourceCommit string) (ReleaseHardeningManifest, string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return ReleaseHardeningManifest{}, "", err
	}
	var matches []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "release-") || strings.HasSuffix(entry.Name(), "-work") {
			continue
		}
		candidate := filepath.Join(root, entry.Name(), "release-hardening-manifest.json")
		if info, statErr := os.Lstat(candidate); statErr == nil && info.Mode().IsRegular() {
			matches = append(matches, candidate)
		}
	}
	if len(matches) != 1 {
		return ReleaseHardeningManifest{}, "", fmt.Errorf("P0/P1 release gate found %d release build manifests, want exactly one", len(matches))
	}
	if err := ValidateReleaseHardeningArtifact(matches[0]); err != nil {
		return ReleaseHardeningManifest{}, "", fmt.Errorf("validate release build evidence: %w", err)
	}
	var manifest ReleaseHardeningManifest
	if err := readUpgradeCoverageRunJSON(matches[0], &manifest); err != nil {
		return ReleaseHardeningManifest{}, "", err
	}
	if manifest.SourceCommit != sourceCommit {
		return ReleaseHardeningManifest{}, "", fmt.Errorf("release build source commit %s, want %s", manifest.SourceCommit, sourceCommit)
	}
	status, err := os.ReadFile(filepath.Join(filepath.Dir(matches[0]), "status.txt"))
	if err != nil {
		return ReleaseHardeningManifest{}, "", err
	}
	if !strings.Contains(string(status), "result=passed\n") || !strings.Contains(string(status), "stage=complete\n") {
		return ReleaseHardeningManifest{}, "", errors.New("release build cleanup status is not an explicit completed pass")
	}
	relative, err := filepath.Rel(root, matches[0])
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return ReleaseHardeningManifest{}, "", errors.New("release build manifest escapes aggregate root")
	}
	return manifest, filepath.ToSlash(relative), nil
}

func discoverP0P1ReleaseGateSuites(root string) ([]P0P1ReleaseGateSuiteEvidence, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	found := make(map[string]P0P1ReleaseGateSuiteEvidence)
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "run-") {
			continue
		}
		runDir := filepath.Join(root, entry.Name())
		var manifest p0p1RunManifest
		if err := readUpgradeCoverageRunJSON(filepath.Join(runDir, "manifest.json"), &manifest); err != nil {
			return nil, fmt.Errorf("read P0/P1 run %s: %w", entry.Name(), err)
		}
		if !isRequiredP0P1ReleaseGateTest(manifest.TestName) {
			return nil, fmt.Errorf("fresh P0/P1 aggregate contains unexpected run %q in %s", manifest.TestName, entry.Name())
		}
		if _, duplicate := found[manifest.TestName]; duplicate {
			return nil, fmt.Errorf("P0/P1 aggregate contains duplicate run for %q", manifest.TestName)
		}
		if err := validateUpgradeCoverageSourceRun(runDir); err != nil {
			return nil, fmt.Errorf("P0/P1 suite %q is not successful and cleaned: %w", manifest.TestName, err)
		}
		finalNodeImages, finalIDs, err := readP0P1FinalNodeImageIDs(runDir, manifest.NumValidators+manifest.NumFullNodes)
		if err != nil {
			return nil, fmt.Errorf("P0/P1 suite %q: %w", manifest.TestName, err)
		}
		_, startsOld := oldImageP0P1ReleaseGateTests[manifest.TestName]
		switches, err := readP0P1SwitchImageIDs(runDir, finalNodeImages, startsOld)
		if err != nil {
			return nil, fmt.Errorf("P0/P1 suite %q: %w", manifest.TestName, err)
		}
		if startsOld {
			if strings.Join(switches.NewImageIDs, ",") != strings.Join(finalIDs, ",") {
				return nil, fmt.Errorf("P0/P1 upgrade suite %q switch target IDs %v differ from final IDs %v", manifest.TestName, switches.NewImageIDs, finalIDs)
			}
		} else if switches.Count != 0 {
			return nil, fmt.Errorf("P0/P1 current-only suite %q unexpectedly switched images", manifest.TestName)
		}
		found[manifest.TestName] = P0P1ReleaseGateSuiteEvidence{
			TestName: manifest.TestName, RunPath: entry.Name(), InitialImageRef: manifest.Image,
			FinalImageIDs: finalIDs, OldImageIDs: switches.OldImageIDs, NodeCount: len(finalNodeImages),
			SwitchCount: switches.Count, SwitchedNodes: switches.Nodes,
		}
	}
	if len(found) != len(requiredP0P1ReleaseGateTests) {
		return nil, fmt.Errorf("P0/P1 aggregate has %d required live suites, want %d", len(found), len(requiredP0P1ReleaseGateTests))
	}
	result := make([]P0P1ReleaseGateSuiteEvidence, 0, len(found))
	for _, testName := range requiredP0P1ReleaseGateTests {
		suite, ok := found[testName]
		if !ok {
			return nil, fmt.Errorf("P0/P1 aggregate is missing live suite %q", testName)
		}
		result = append(result, suite)
	}
	return result, nil
}

func isRequiredP0P1ReleaseGateTest(testName string) bool {
	for _, required := range requiredP0P1ReleaseGateTests {
		if testName == required {
			return true
		}
	}
	return false
}

func readP0P1FinalNodeImageIDs(runDir string, expectedNodes int) (map[string]string, []string, error) {
	if expectedNodes < 1 {
		return nil, nil, fmt.Errorf("run declares %d Panacea nodes", expectedNodes)
	}
	paths, err := filepath.Glob(filepath.Join(runDir, "nodes", "*", "container-state.json"))
	if err != nil {
		return nil, nil, err
	}
	if len(paths) != expectedNodes {
		return nil, nil, fmt.Errorf("found %d final node container states, want %d", len(paths), expectedNodes)
	}
	ids := make(map[string]struct{})
	nodes := make(map[string]string, len(paths))
	for _, filePath := range paths {
		var state p0p1ContainerState
		if err := readUpgradeCoverageRunJSON(filePath, &state); err != nil {
			return nil, nil, err
		}
		if !releaseDigestPattern.MatchString(state.Image) {
			return nil, nil, fmt.Errorf("node container image ID %q is not sha256", state.Image)
		}
		node := filepath.Base(filepath.Dir(filePath))
		if strings.TrimSpace(node) == "" {
			return nil, nil, errors.New("final node container state has an empty node name")
		}
		nodes[node] = state.Image
		ids[state.Image] = struct{}{}
	}
	return nodes, sortedP0P1Values(ids), nil
}

type p0p1SwitchSummary struct {
	OldImageIDs []string
	NewImageIDs []string
	Nodes       []string
	Count       int
}

func readP0P1SwitchImageIDs(runDir string, finalNodeImages map[string]string, requireAll bool) (p0p1SwitchSummary, error) {
	filePath := filepath.Join(runDir, "upgrade", "node-switches.jsonl")
	file, err := os.Open(filePath)
	if errors.Is(err, fs.ErrNotExist) {
		if requireAll {
			return p0p1SwitchSummary{}, fmt.Errorf("image switch evidence is missing for %d declared nodes", len(finalNodeImages))
		}
		return p0p1SwitchSummary{}, nil
	}
	if err != nil {
		return p0p1SwitchSummary{}, err
	}
	defer file.Close()
	oldIDs := make(map[string]struct{})
	newIDs := make(map[string]struct{})
	nodes := make(map[string]struct{}, len(finalNodeImages))
	count := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64<<10), upgradeCoverageMatrixReadLimit)
	for scanner.Scan() {
		var record p0p1SwitchRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return p0p1SwitchSummary{}, fmt.Errorf("decode image switch line %d: %w", count+1, err)
		}
		if record.Error != "" || !releaseDigestPattern.MatchString(record.OldImageID) || !releaseDigestPattern.MatchString(record.NewImageID) || record.OldImageID == record.NewImageID {
			return p0p1SwitchSummary{}, fmt.Errorf("image switch line %d is not a successful distinct sha256 transition", count+1)
		}
		node := strings.TrimSpace(record.Plan.Node)
		finalImageID, declared := finalNodeImages[node]
		if !declared {
			return p0p1SwitchSummary{}, fmt.Errorf("image switch line %d names undeclared node %q", count+1, node)
		}
		if _, duplicate := nodes[node]; duplicate {
			return p0p1SwitchSummary{}, fmt.Errorf("image switch node %q appears more than once", node)
		}
		if record.NewImageID != finalImageID {
			return p0p1SwitchSummary{}, fmt.Errorf("image switch node %q target %s differs from final image %s", node, record.NewImageID, finalImageID)
		}
		oldIDs[record.OldImageID] = struct{}{}
		newIDs[record.NewImageID] = struct{}{}
		nodes[node] = struct{}{}
		count++
	}
	if err := scanner.Err(); err != nil {
		return p0p1SwitchSummary{}, err
	}
	if requireAll && count != len(finalNodeImages) {
		return p0p1SwitchSummary{}, fmt.Errorf("found %d unique node image switches, want exactly %d declared nodes", count, len(finalNodeImages))
	}
	return p0p1SwitchSummary{
		OldImageIDs: sortedP0P1Values(oldIDs), NewImageIDs: sortedP0P1Values(newIDs),
		Nodes: sortedP0P1Values(nodes), Count: count,
	}, nil
}

func requireOneP0P1Value(values map[string]struct{}, label string) (string, error) {
	if len(values) != 1 {
		return "", fmt.Errorf("P0/P1 release gate found %d distinct %s values, want one: %v", len(values), label, sortedP0P1Values(values))
	}
	for value := range values {
		return value, nil
	}
	panic("unreachable")
}

func sortedP0P1Values(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func writeP0P1ReleaseGateJSON(root, relativePath string, value any) error {
	if err := validateCoverageArtifactPath(relativePath); err != nil {
		return err
	}
	output := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := rejectUpgradeCoverageSymlinkComponents(root, filepath.Dir(output)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o700); err != nil {
		return err
	}
	if err := rejectUpgradeCoverageSymlinkComponents(root, filepath.Dir(output)); err != nil {
		return err
	}
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	contents = append(contents, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(output), ".release-gate-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, output); err != nil {
		return err
	}
	return nil
}
