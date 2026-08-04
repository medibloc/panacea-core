package harness

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
)

const (
	UpgradeCoverageMatrixSchemaVersion = "4"
	UpgradeCoverageMatrixArtifactPath  = "upgrade/coverage-matrix.json"
	UpgradeQueryCoverageArtifactPath   = "upgrade/query-coverage.json"
)

type UpgradeQueryBoundary string

const (
	UpgradeQueryBoundaryCLI  UpgradeQueryBoundary = "CLI"
	UpgradeQueryBoundaryGRPC UpgradeQueryBoundary = "gRPC"
	UpgradeQueryBoundaryREST UpgradeQueryBoundary = "REST"
)

var requiredUpgradeQueryBoundaries = []UpgradeQueryBoundary{
	UpgradeQueryBoundaryCLI,
	UpgradeQueryBoundaryGRPC,
	UpgradeQueryBoundaryREST,
}

type UpgradeCoveragePriority string

const (
	UpgradeCoveragePriorityP0 UpgradeCoveragePriority = "P0"
	UpgradeCoveragePriorityP1 UpgradeCoveragePriority = "P1"
)

type UpgradeCoverageArea string

const (
	UpgradeCoverageAreaAuthBank      UpgradeCoverageArea = "auth/bank"
	UpgradeCoverageAreaGov           UpgradeCoverageArea = "gov"
	UpgradeCoverageAreaStaking       UpgradeCoverageArea = "staking"
	UpgradeCoverageAreaDistribution  UpgradeCoverageArea = "distribution"
	UpgradeCoverageAreaSlashing      UpgradeCoverageArea = "slashing"
	UpgradeCoverageAreaDID           UpgradeCoverageArea = "DID"
	UpgradeCoverageAreaAOL           UpgradeCoverageArea = "AOL"
	UpgradeCoverageAreaNFT           UpgradeCoverageArea = "NFT"
	UpgradeCoverageAreaLegacyPNFT    UpgradeCoverageArea = "legacy PNFT"
	UpgradeCoverageAreaAuthzFeegrant UpgradeCoverageArea = "authz/feegrant"
	UpgradeCoverageAreaGroupVesting  UpgradeCoverageArea = "group/vesting"
	UpgradeCoverageAreaSystemModules UpgradeCoverageArea = "system modules"
	UpgradeCoverageAreaIBCTransfer   UpgradeCoverageArea = "IBC/transfer"
)

type upgradeCoverageAreaRequirement struct {
	Area     UpgradeCoverageArea
	Priority UpgradeCoveragePriority
}

var upgradeCoverageAreaRequirements = []upgradeCoverageAreaRequirement{
	{Area: UpgradeCoverageAreaAuthBank, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaGov, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaStaking, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaDistribution, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaSlashing, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaDID, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaAOL, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaNFT, Priority: UpgradeCoveragePriorityP0},
	{Area: UpgradeCoverageAreaLegacyPNFT, Priority: UpgradeCoveragePriorityP1},
	{Area: UpgradeCoverageAreaAuthzFeegrant, Priority: UpgradeCoveragePriorityP1},
	{Area: UpgradeCoverageAreaGroupVesting, Priority: UpgradeCoveragePriorityP1},
	{Area: UpgradeCoverageAreaSystemModules, Priority: UpgradeCoveragePriorityP1},
	{Area: UpgradeCoverageAreaIBCTransfer, Priority: UpgradeCoveragePriorityP1},
}

type UpgradeCoveragePhaseName string

const (
	UpgradeCoveragePhaseV221Preparation         UpgradeCoveragePhaseName = "v2.2.1-preparation"
	UpgradeCoveragePhasePreUpgradeCheckpoint    UpgradeCoveragePhaseName = "pre-upgrade-checkpoint"
	UpgradeCoveragePhasePostUpgradePreservation UpgradeCoveragePhaseName = "post-upgrade-preservation"
	UpgradeCoveragePhasePostUpgradeMutation     UpgradeCoveragePhaseName = "post-upgrade-mutation"
	UpgradeCoveragePhasePostRestart             UpgradeCoveragePhaseName = "post-restart"
)

var requiredUpgradeCoveragePhases = []UpgradeCoveragePhaseName{
	UpgradeCoveragePhaseV221Preparation,
	UpgradeCoveragePhasePreUpgradeCheckpoint,
	UpgradeCoveragePhasePostUpgradePreservation,
	UpgradeCoveragePhasePostUpgradeMutation,
	UpgradeCoveragePhasePostRestart,
}

type UpgradeCoverageStatus string

const (
	UpgradeCoverageStatusPassed  UpgradeCoverageStatus = "passed"
	UpgradeCoverageStatusFailed  UpgradeCoverageStatus = "failed"
	UpgradeCoverageStatusSkipped UpgradeCoverageStatus = "skipped"
	UpgradeCoverageStatusNotRun  UpgradeCoverageStatus = "not-run"
)

// UpgradeCoverageMatrix is the durable audit contract for every required
// transaction/state area across the five connected upgrade phases.
type UpgradeCoverageMatrix struct {
	SchemaVersion  string               `json:"schema_version"`
	RecordedAt     time.Time            `json:"recorded_at"`
	UpgradeName    string               `json:"upgrade_name"`
	SourceVersion  string               `json:"source_version"`
	TargetVersion  string               `json:"target_version"`
	SourceMatrices []string             `json:"source_matrices,omitempty"`
	Rows           []UpgradeCoverageRow `json:"rows"`
}

type UpgradeCoverageRow struct {
	Area            UpgradeCoverageArea     `json:"area"`
	Priority        UpgradeCoveragePriority `json:"priority"`
	Status          UpgradeCoverageStatus   `json:"status"`
	StateObjectIDs  []string                `json:"state_object_ids"`
	LowerLevelTests []string                `json:"lower_level_tests,omitempty"`
	QueryCoverage   []UpgradeQueryCoverage  `json:"query_coverage"`
	Phases          []UpgradeCoveragePhase  `json:"phases"`
}

// UpgradeQueryCoverage separates a boundary's declared capability from the
// live evidence this lane actually exercised. Historical capability and
// server-validated historical execution are also separate, so a successful
// request carrying --height cannot by itself prove which height served it.
type UpgradeQueryCoverage struct {
	Boundary                      UpgradeQueryBoundary            `json:"boundary"`
	Supported                     bool                            `json:"supported"`
	Exercised                     bool                            `json:"exercised"`
	Reason                        string                          `json:"reason"`
	EvidencePaths                 []string                        `json:"evidence_paths"`
	Evidence                      []UpgradeQueryEvidenceReference `json:"evidence,omitempty"`
	HistoricalHeightSupported     bool                            `json:"historical_height_supported"`
	HistoricalHeightExercised     bool                            `json:"historical_height_exercised"`
	HistoricalHeightReason        string                          `json:"historical_height_reason"`
	HistoricalHeightEvidencePaths []string                        `json:"historical_height_evidence_paths"`
	HistoricalHeightEvidence      []UpgradeQueryEvidenceReference `json:"historical_height_evidence,omitempty"`
}

// UpgradeQueryEvidenceReference selects one successful transport record from
// a JSONL query artifact. Boundary and step are exact matches, while
// HistoricalHeight declares whether the selected request carries a positive
// application height. Only HistoricalHeightEvidence is additionally required
// to prove that the server returned that exact height.
type UpgradeQueryEvidenceReference struct {
	ArtifactPath     string               `json:"artifact_path"`
	Boundary         UpgradeQueryBoundary `json:"boundary"`
	Step             string               `json:"step"`
	HistoricalHeight bool                 `json:"historical_height"`
}

type UpgradeQueryCoverageContract struct {
	SchemaVersion string                    `json:"schema_version"`
	RecordedAt    time.Time                 `json:"recorded_at"`
	UpgradeName   string                    `json:"upgrade_name"`
	SourceVersion string                    `json:"source_version"`
	TargetVersion string                    `json:"target_version"`
	Rows          []UpgradeQueryCoverageRow `json:"rows"`
}

type UpgradeQueryCoverageRow struct {
	Area           UpgradeCoverageArea    `json:"area"`
	StateObjectIDs []string               `json:"state_object_ids"`
	Boundaries     []UpgradeQueryCoverage `json:"boundaries"`
}

type UpgradeCoveragePhase struct {
	Name          UpgradeCoveragePhaseName `json:"name"`
	Status        UpgradeCoverageStatus    `json:"status"`
	ArtifactPaths []string                 `json:"artifact_paths,omitempty"`
	Reason        string                   `json:"reason,omitempty"`
}

// Validate rejects incomplete matrices while still allowing failed, skipped,
// and not-run phases to be represented with an explicit reason. This ensures a
// failed run can report coverage gaps instead of silently omitting rows.
func (m UpgradeCoverageMatrix) Validate() error {
	if m.SchemaVersion != UpgradeCoverageMatrixSchemaVersion {
		return fmt.Errorf("coverage matrix schema version %q, want %q", m.SchemaVersion, UpgradeCoverageMatrixSchemaVersion)
	}
	if m.RecordedAt.IsZero() {
		return errors.New("coverage matrix recorded_at is required")
	}
	if strings.TrimSpace(m.UpgradeName) == "" {
		return errors.New("coverage matrix upgrade_name is required")
	}
	if strings.TrimSpace(m.SourceVersion) == "" || strings.TrimSpace(m.TargetVersion) == "" {
		return errors.New("coverage matrix source_version and target_version are required")
	}
	if m.SourceVersion == m.TargetVersion {
		return errors.New("coverage matrix source_version and target_version must differ")
	}
	if err := validateUniqueNonEmptyStrings(m.SourceMatrices, "source_matrices"); err != nil {
		return err
	}
	for _, sourceMatrix := range m.SourceMatrices {
		if err := validateCoverageArtifactPath(sourceMatrix); err != nil {
			return fmt.Errorf("source matrix: %w", err)
		}
	}
	if len(m.Rows) != len(upgradeCoverageAreaRequirements) {
		return fmt.Errorf("coverage matrix has %d rows, want %d", len(m.Rows), len(upgradeCoverageAreaRequirements))
	}

	requirements := make(map[UpgradeCoverageArea]UpgradeCoveragePriority, len(upgradeCoverageAreaRequirements))
	for _, requirement := range upgradeCoverageAreaRequirements {
		requirements[requirement.Area] = requirement.Priority
	}
	seenAreas := make(map[UpgradeCoverageArea]struct{}, len(m.Rows))
	for index, row := range m.Rows {
		expectedPriority, required := requirements[row.Area]
		if !required {
			return fmt.Errorf("coverage matrix row %d has unknown area %q", index, row.Area)
		}
		if _, duplicate := seenAreas[row.Area]; duplicate {
			return fmt.Errorf("coverage matrix contains duplicate area %q", row.Area)
		}
		seenAreas[row.Area] = struct{}{}
		if row.Priority != expectedPriority {
			return fmt.Errorf("coverage matrix area %q priority %q, want %q", row.Area, row.Priority, expectedPriority)
		}
		if err := validateUpgradeCoverageRow(row); err != nil {
			return fmt.Errorf("coverage matrix area %q: %w", row.Area, err)
		}
	}
	for area := range requirements {
		if _, present := seenAreas[area]; !present {
			return fmt.Errorf("coverage matrix is missing required area %q", area)
		}
	}
	return nil
}

func validateUpgradeCoverageRow(row UpgradeCoverageRow) error {
	if len(row.StateObjectIDs) == 0 {
		return errors.New("state_object_ids must not be empty")
	}
	if err := validateUniqueNonEmptyStrings(row.StateObjectIDs, "state_object_ids"); err != nil {
		return err
	}
	if err := validateUniqueNonEmptyStrings(row.LowerLevelTests, "lower_level_tests"); err != nil {
		return err
	}
	if err := validateUpgradeQueryCoverage(row.QueryCoverage); err != nil {
		return err
	}
	if len(row.Phases) != len(requiredUpgradeCoveragePhases) {
		return fmt.Errorf("has %d phases, want %d", len(row.Phases), len(requiredUpgradeCoveragePhases))
	}
	for index, expectedName := range requiredUpgradeCoveragePhases {
		phase := row.Phases[index]
		if phase.Name != expectedName {
			return fmt.Errorf("phase %d name %q, want %q", index, phase.Name, expectedName)
		}
		if err := validateUpgradeCoveragePhase(phase); err != nil {
			return fmt.Errorf("phase %q: %w", phase.Name, err)
		}
	}
	expectedStatus, err := aggregateUpgradeCoverageStatus(row.Phases)
	if err != nil {
		return err
	}
	if row.Status != expectedStatus {
		return fmt.Errorf("row status %q, want aggregate phase status %q", row.Status, expectedStatus)
	}
	return nil
}

func validateUpgradeQueryCoverage(coverage []UpgradeQueryCoverage) error {
	if len(coverage) != len(requiredUpgradeQueryBoundaries) {
		return fmt.Errorf("query_coverage has %d boundaries, want %d", len(coverage), len(requiredUpgradeQueryBoundaries))
	}
	for index, expectedBoundary := range requiredUpgradeQueryBoundaries {
		claim := coverage[index]
		if claim.Boundary != expectedBoundary {
			return fmt.Errorf("query_coverage boundary %d is %q, want %q", index, claim.Boundary, expectedBoundary)
		}
		if strings.TrimSpace(claim.Reason) == "" {
			return fmt.Errorf("query_coverage boundary %q reason is required", claim.Boundary)
		}
		if strings.TrimSpace(claim.HistoricalHeightReason) == "" {
			return fmt.Errorf("query_coverage boundary %q historical_height_reason is required", claim.Boundary)
		}
		if claim.Exercised && !claim.Supported {
			return fmt.Errorf("query_coverage boundary %q cannot be exercised while unsupported", claim.Boundary)
		}
		if claim.HistoricalHeightSupported && !claim.Supported {
			return fmt.Errorf("query_coverage boundary %q cannot support historical height while the boundary is unsupported", claim.Boundary)
		}
		if claim.HistoricalHeightExercised && !claim.HistoricalHeightSupported {
			return fmt.Errorf("query_coverage boundary %q cannot exercise historical height while it is unsupported", claim.Boundary)
		}
		if claim.HistoricalHeightExercised && !claim.Exercised {
			return fmt.Errorf("query_coverage boundary %q cannot exercise historical height without exercising the boundary", claim.Boundary)
		}
		if err := validateUpgradeQueryEvidencePaths(claim.EvidencePaths, claim.Exercised, "query", claim.Boundary); err != nil {
			return err
		}
		if err := validateUpgradeQueryEvidenceReferences(
			claim.Evidence,
			claim.EvidencePaths,
			claim.Exercised,
			false,
			"query",
			claim.Boundary,
		); err != nil {
			return err
		}
		if err := validateUpgradeQueryEvidencePaths(
			claim.HistoricalHeightEvidencePaths,
			claim.HistoricalHeightExercised,
			"historical-height query",
			claim.Boundary,
		); err != nil {
			return err
		}
		if err := validateUpgradeQueryEvidenceReferences(
			claim.HistoricalHeightEvidence,
			claim.HistoricalHeightEvidencePaths,
			claim.HistoricalHeightExercised,
			true,
			"historical-height query",
			claim.Boundary,
		); err != nil {
			return err
		}
	}
	return nil
}

func validateUpgradeQueryEvidenceReferences(
	references []UpgradeQueryEvidenceReference,
	paths []string,
	exercised bool,
	requireHistorical bool,
	kind string,
	boundary UpgradeQueryBoundary,
) error {
	if !exercised {
		if len(references) != 0 {
			return fmt.Errorf("query_coverage boundary %q unexercised %s claim must not include query evidence references", boundary, kind)
		}
		return nil
	}
	if len(references) == 0 {
		return fmt.Errorf("query_coverage boundary %q exercised %s claim requires a query evidence reference", boundary, kind)
	}
	pathSet := make(map[string]struct{}, len(paths))
	for _, artifactPath := range paths {
		pathSet[artifactPath] = struct{}{}
	}
	referencedPaths := make(map[string]struct{}, len(references))
	seen := make(map[string]struct{}, len(references))
	for _, reference := range references {
		if reference.Boundary != boundary {
			return fmt.Errorf("query_coverage boundary %q %s evidence declares boundary %q", boundary, kind, reference.Boundary)
		}
		if strings.TrimSpace(reference.Step) == "" {
			return fmt.Errorf("query_coverage boundary %q %s evidence step is required", boundary, kind)
		}
		if err := validateCoverageArtifactPath(reference.ArtifactPath); err != nil {
			return fmt.Errorf("query_coverage boundary %q %s evidence: %w", boundary, kind, err)
		}
		if _, present := pathSet[reference.ArtifactPath]; !present {
			return fmt.Errorf("query_coverage boundary %q %s evidence path %q is absent from evidence_paths", boundary, kind, reference.ArtifactPath)
		}
		if requireHistorical && !reference.HistoricalHeight {
			return fmt.Errorf("query_coverage boundary %q historical evidence step %q is not height-pinned", boundary, reference.Step)
		}
		key := reference.ArtifactPath + "\x00" + string(reference.Boundary) + "\x00" + reference.Step + "\x00" + fmt.Sprint(reference.HistoricalHeight)
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("query_coverage boundary %q contains duplicate %s evidence step %q", boundary, kind, reference.Step)
		}
		seen[key] = struct{}{}
		referencedPaths[reference.ArtifactPath] = struct{}{}
	}
	if len(referencedPaths) != len(pathSet) {
		return fmt.Errorf("query_coverage boundary %q %s evidence_paths must identify only structured query artifacts", boundary, kind)
	}
	return nil
}

func validateUpgradeQueryEvidencePaths(paths []string, exercised bool, kind string, boundary UpgradeQueryBoundary) error {
	if len(paths) == 0 {
		return fmt.Errorf("query_coverage boundary %q %s evidence_paths must not be empty", boundary, kind)
	}
	seen := make(map[string]struct{}, len(paths))
	hasLiveEvidence := false
	for _, artifactPath := range paths {
		if err := validateCoverageArtifactPath(artifactPath); err != nil {
			return fmt.Errorf("query_coverage boundary %q: %w", boundary, err)
		}
		if _, duplicate := seen[artifactPath]; duplicate {
			return fmt.Errorf("query_coverage boundary %q has duplicate %s evidence path %q", boundary, kind, artifactPath)
		}
		seen[artifactPath] = struct{}{}
		if artifactPath != UpgradeQueryCoverageArtifactPath {
			hasLiveEvidence = true
		}
	}
	if exercised && !hasLiveEvidence {
		return fmt.Errorf("query_coverage boundary %q exercised %s claim requires live evidence beyond %s", boundary, kind, UpgradeQueryCoverageArtifactPath)
	}
	return nil
}

func validateUpgradeCoveragePhase(phase UpgradeCoveragePhase) error {
	switch phase.Status {
	case UpgradeCoverageStatusPassed:
		if len(phase.ArtifactPaths) == 0 {
			return errors.New("passed phase must reference at least one artifact")
		}
	case UpgradeCoverageStatusFailed, UpgradeCoverageStatusSkipped, UpgradeCoverageStatusNotRun:
		if strings.TrimSpace(phase.Reason) == "" {
			return fmt.Errorf("%s phase must include a reason", phase.Status)
		}
	default:
		return fmt.Errorf("unknown status %q", phase.Status)
	}
	seenPaths := make(map[string]struct{}, len(phase.ArtifactPaths))
	for _, artifactPath := range phase.ArtifactPaths {
		if err := validateCoverageArtifactPath(artifactPath); err != nil {
			return err
		}
		if _, duplicate := seenPaths[artifactPath]; duplicate {
			return fmt.Errorf("duplicate artifact path %q", artifactPath)
		}
		seenPaths[artifactPath] = struct{}{}
	}
	return nil
}

func aggregateUpgradeCoverageStatus(phases []UpgradeCoveragePhase) (UpgradeCoverageStatus, error) {
	status := UpgradeCoverageStatusPassed
	for _, phase := range phases {
		switch phase.Status {
		case UpgradeCoverageStatusFailed:
			return UpgradeCoverageStatusFailed, nil
		case UpgradeCoverageStatusNotRun:
			status = UpgradeCoverageStatusNotRun
		case UpgradeCoverageStatusSkipped:
			if status == UpgradeCoverageStatusPassed {
				status = UpgradeCoverageStatusSkipped
			}
		case UpgradeCoverageStatusPassed:
		default:
			return "", fmt.Errorf("phase %q has unknown status %q", phase.Name, phase.Status)
		}
	}
	return status, nil
}

func validateUniqueNonEmptyStrings(values []string, field string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("%s contains an empty value", field)
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("%s contains duplicate value %q", field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateCoverageArtifactPath(artifactPath string) error {
	if strings.TrimSpace(artifactPath) == "" {
		return errors.New("artifact path must not be empty")
	}
	if path.IsAbs(artifactPath) {
		return fmt.Errorf("artifact path must be relative: %q", artifactPath)
	}
	clean := path.Clean(artifactPath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || clean != artifactPath {
		return fmt.Errorf("artifact path must be a clean run-relative path: %q", artifactPath)
	}
	return nil
}

// RecordUpgradeCoverageMatrix validates and writes the complete matrix to its
// stable artifact path. Invalid matrices are never partially recorded.
func (n *Network) RecordUpgradeCoverageMatrix(matrix UpgradeCoverageMatrix) error {
	if n == nil || n.artifacts == nil {
		return errors.New("upgrade coverage matrix artifact store is unavailable")
	}
	if err := matrix.Validate(); err != nil {
		return fmt.Errorf("validate upgrade coverage matrix: %w", err)
	}
	contract := UpgradeQueryCoverageContract{
		SchemaVersion: matrix.SchemaVersion,
		RecordedAt:    matrix.RecordedAt,
		UpgradeName:   matrix.UpgradeName,
		SourceVersion: matrix.SourceVersion,
		TargetVersion: matrix.TargetVersion,
		Rows:          make([]UpgradeQueryCoverageRow, 0, len(matrix.Rows)),
	}
	for _, row := range matrix.Rows {
		contract.Rows = append(contract.Rows, UpgradeQueryCoverageRow{
			Area:           row.Area,
			StateObjectIDs: append([]string(nil), row.StateObjectIDs...),
			Boundaries:     append([]UpgradeQueryCoverage(nil), row.QueryCoverage...),
		})
	}
	if err := n.artifacts.writeJSON(UpgradeQueryCoverageArtifactPath, contract); err != nil {
		return fmt.Errorf("record upgrade query coverage contract: %w", err)
	}
	if err := n.artifacts.writeJSON(UpgradeCoverageMatrixArtifactPath, matrix); err != nil {
		return fmt.Errorf("record upgrade coverage matrix: %w", err)
	}
	return nil
}
