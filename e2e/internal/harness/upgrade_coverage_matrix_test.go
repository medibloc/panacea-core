package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpgradeCoverageMatrixValidate(t *testing.T) {
	matrix := validUpgradeCoverageMatrix()
	require.NoError(t, matrix.Validate())
}

func TestUpgradeCoverageMatrixSeparatesSupportedFromExercised(t *testing.T) {
	matrix := validUpgradeCoverageMatrix()
	claim := &matrix.Rows[0].QueryCoverage[1]
	claim.Supported = true
	claim.Exercised = false
	claim.Reason = "gRPC is part of the module API, but this lane has no structured live record"

	require.NoError(t, matrix.Validate())
}

func TestUpgradeCoverageMatrixAllowsExplicitFailureAndUnrunPhases(t *testing.T) {
	matrix := validUpgradeCoverageMatrix()
	row := &matrix.Rows[0]
	row.Phases[2] = UpgradeCoveragePhase{
		Name:   UpgradeCoveragePhasePostUpgradePreservation,
		Status: UpgradeCoverageStatusFailed,
		Reason: "query returned an unexpected account sequence",
	}
	row.Phases[3] = UpgradeCoveragePhase{
		Name:   UpgradeCoveragePhasePostUpgradeMutation,
		Status: UpgradeCoverageStatusNotRun,
		Reason: "blocked by preservation failure",
	}
	row.Phases[4] = UpgradeCoveragePhase{
		Name:   UpgradeCoveragePhasePostRestart,
		Status: UpgradeCoverageStatusNotRun,
		Reason: "blocked by preservation failure",
	}
	row.Status = UpgradeCoverageStatusFailed

	require.NoError(t, matrix.Validate())
}

func TestUpgradeCoverageMatrixRejectsInvalidContracts(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*UpgradeCoverageMatrix)
		wantError string
	}{
		{
			name: "missing required row",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows = matrix.Rows[:len(matrix.Rows)-1]
			},
			wantError: "rows, want",
		},
		{
			name: "duplicate area",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[1].Area = matrix.Rows[0].Area
				matrix.Rows[1].Priority = matrix.Rows[0].Priority
			},
			wantError: "duplicate area",
		},
		{
			name: "wrong priority",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].Priority = UpgradeCoveragePriorityP1
			},
			wantError: "priority \"P1\", want \"P0\"",
		},
		{
			name: "missing object identity",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].StateObjectIDs = nil
			},
			wantError: "state_object_ids must not be empty",
		},
		{
			name: "missing query boundary",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].QueryCoverage = matrix.Rows[0].QueryCoverage[:2]
			},
			wantError: "query_coverage has 2 boundaries, want 3",
		},
		{
			name: "query cannot be exercised while unsupported",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].QueryCoverage[0].Supported = false
				matrix.Rows[0].QueryCoverage[0].HistoricalHeightSupported = false
			},
			wantError: "cannot be exercised while unsupported",
		},
		{
			name: "historical query cannot exceed boundary support",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].QueryCoverage[0].Supported = false
				matrix.Rows[0].QueryCoverage[0].Exercised = false
			},
			wantError: "cannot support historical height",
		},
		{
			name: "exercised query requires live evidence",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].QueryCoverage[0].EvidencePaths = []string{UpgradeQueryCoverageArtifactPath}
			},
			wantError: "requires live evidence",
		},
		{
			name: "exercised query requires structured reference",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].QueryCoverage[0].Evidence = nil
			},
			wantError: "requires a query evidence reference",
		},
		{
			name: "query evidence boundary must match claim",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].QueryCoverage[0].Evidence[0].Boundary = UpgradeQueryBoundaryREST
			},
			wantError: "declares boundary",
		},
		{
			name: "historical query evidence must declare pin",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				claim := &matrix.Rows[0].QueryCoverage[0]
				claim.HistoricalHeightExercised = true
				claim.HistoricalHeightEvidencePaths = []string{"queries/results.jsonl"}
				claim.HistoricalHeightEvidence = []UpgradeQueryEvidenceReference{{
					ArtifactPath: "queries/results.jsonl",
					Boundary:     UpgradeQueryBoundaryCLI,
					Step:         "fixture-historical-cli",
				}}
				matrix.Rows[0].QueryCoverage[0].HistoricalHeightEvidence[0].HistoricalHeight = false
			},
			wantError: "is not height-pinned",
		},
		{
			name: "missing phase",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].Phases = matrix.Rows[0].Phases[:4]
			},
			wantError: "has 4 phases, want 5",
		},
		{
			name: "phase order changed",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].Phases[0], matrix.Rows[0].Phases[1] = matrix.Rows[0].Phases[1], matrix.Rows[0].Phases[0]
			},
			wantError: "phase 0 name",
		},
		{
			name: "passed phase has no evidence",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].Phases[0].ArtifactPaths = nil
			},
			wantError: "passed phase must reference",
		},
		{
			name: "skipped phase has no reason",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].Phases[0].Status = UpgradeCoverageStatusSkipped
				matrix.Rows[0].Phases[0].ArtifactPaths = nil
				matrix.Rows[0].Status = UpgradeCoverageStatusSkipped
			},
			wantError: "skipped phase must include a reason",
		},
		{
			name: "row status disagrees with phases",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].Status = UpgradeCoverageStatusFailed
			},
			wantError: "row status \"failed\", want aggregate phase status \"passed\"",
		},
		{
			name: "artifact path escapes run root",
			mutate: func(matrix *UpgradeCoverageMatrix) {
				matrix.Rows[0].Phases[0].ArtifactPaths = []string{"../outside.json"}
			},
			wantError: "clean run-relative path",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matrix := validUpgradeCoverageMatrix()
			test.mutate(&matrix)
			err := matrix.Validate()
			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestRecordUpgradeCoverageMatrix(t *testing.T) {
	store, err := newArtifactStore(
		"upgrade-coverage-matrix",
		"run-coverage",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	require.NoError(t, err)
	network := &Network{artifacts: store}
	matrix := validUpgradeCoverageMatrix()

	require.NoError(t, network.RecordUpgradeCoverageMatrix(matrix))
	contents, err := os.ReadFile(filepath.Join(store.dir, filepath.FromSlash(UpgradeCoverageMatrixArtifactPath)))
	require.NoError(t, err)
	var recorded UpgradeCoverageMatrix
	require.NoError(t, json.Unmarshal(contents, &recorded))
	require.Equal(t, matrix, recorded)
	queryContents, err := os.ReadFile(filepath.Join(store.dir, filepath.FromSlash(UpgradeQueryCoverageArtifactPath)))
	require.NoError(t, err)
	var queryContract UpgradeQueryCoverageContract
	require.NoError(t, json.Unmarshal(queryContents, &queryContract))
	require.Len(t, queryContract.Rows, len(matrix.Rows))
	require.Equal(t, matrix.Rows[0].QueryCoverage, queryContract.Rows[0].Boundaries)
}

func TestRecordUpgradeCoverageMatrixDoesNotWriteInvalidMatrix(t *testing.T) {
	store, err := newArtifactStore(
		"upgrade-coverage-matrix-invalid",
		"run-coverage-invalid",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	require.NoError(t, err)
	network := &Network{artifacts: store}
	matrix := validUpgradeCoverageMatrix()
	matrix.Rows[0].Phases = nil

	err = network.RecordUpgradeCoverageMatrix(matrix)
	require.ErrorContains(t, err, "validate upgrade coverage matrix")
	_, statErr := os.Stat(filepath.Join(store.dir, filepath.FromSlash(UpgradeCoverageMatrixArtifactPath)))
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func validUpgradeCoverageMatrix() UpgradeCoverageMatrix {
	rows := make([]UpgradeCoverageRow, 0, len(upgradeCoverageAreaRequirements))
	for _, requirement := range upgradeCoverageAreaRequirements {
		slug := strings.NewReplacer("/", "-", " ", "-").Replace(strings.ToLower(string(requirement.Area)))
		phases := make([]UpgradeCoveragePhase, 0, len(requiredUpgradeCoveragePhases))
		for _, phase := range requiredUpgradeCoveragePhases {
			phases = append(phases, UpgradeCoveragePhase{
				Name:          phase,
				Status:        UpgradeCoverageStatusPassed,
				ArtifactPaths: []string{"state-checkpoints/" + slug + "/" + string(phase) + ".json"},
			})
		}
		rows = append(rows, UpgradeCoverageRow{
			Area:           requirement.Area,
			Priority:       requirement.Priority,
			Status:         UpgradeCoverageStatusPassed,
			StateObjectIDs: []string{string(requirement.Area) + ":fixture"},
			QueryCoverage: []UpgradeQueryCoverage{
				{
					Boundary:      UpgradeQueryBoundaryCLI,
					Supported:     true,
					Exercised:     true,
					Reason:        "CLI latest-state query exercised by the fixture",
					EvidencePaths: []string{"queries/results.jsonl"},
					Evidence: []UpgradeQueryEvidenceReference{{
						ArtifactPath:     "queries/results.jsonl",
						Boundary:         UpgradeQueryBoundaryCLI,
						Step:             "fixture-" + slug + "-cli",
						HistoricalHeight: true,
					}},
					HistoricalHeightSupported:     true,
					HistoricalHeightExercised:     false,
					HistoricalHeightReason:        "CLI request is height-pinned but has no server-returned response height",
					HistoricalHeightEvidencePaths: []string{UpgradeQueryCoverageArtifactPath},
				},
				{
					Boundary:                      UpgradeQueryBoundaryGRPC,
					Supported:                     false,
					Reason:                        "this fixture does not exercise gRPC",
					EvidencePaths:                 []string{UpgradeQueryCoverageArtifactPath},
					HistoricalHeightSupported:     false,
					HistoricalHeightReason:        "this fixture does not exercise height-pinned gRPC",
					HistoricalHeightEvidencePaths: []string{UpgradeQueryCoverageArtifactPath},
				},
				{
					Boundary:                      UpgradeQueryBoundaryREST,
					Supported:                     false,
					Reason:                        "this fixture does not exercise REST",
					EvidencePaths:                 []string{UpgradeQueryCoverageArtifactPath},
					HistoricalHeightSupported:     false,
					HistoricalHeightReason:        "this fixture does not exercise height-pinned REST",
					HistoricalHeightEvidencePaths: []string{UpgradeQueryCoverageArtifactPath},
				},
			},
			Phases: phases,
		})
	}
	return UpgradeCoverageMatrix{
		SchemaVersion: UpgradeCoverageMatrixSchemaVersion,
		RecordedAt:    time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
		UpgradeName:   "v2.3.0",
		SourceVersion: "v2.2.1",
		TargetVersion: "v2.3.0",
		Rows:          rows,
	}
}
