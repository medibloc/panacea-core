package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMergeUpgradeCoverageMatricesRequiresRealArtifactsAndClosesEveryGap(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	connected := validUpgradeCoverageMatrix()
	setCoverageRowNotRun(&connected, UpgradeCoverageAreaIBCTransfer, "isolated IBC lane")
	ibc := validUpgradeCoverageMatrix()
	for _, requirement := range upgradeCoverageAreaRequirements {
		if requirement.Area != UpgradeCoverageAreaIBCTransfer {
			setCoverageRowNotRun(&ibc, requirement.Area, "connected upgrade lane")
		}
	}
	writeCoverageFixture(t, root, "run-connected", connected)
	writeCoverageFixture(t, root, "run-ibc", ibc)

	merged, err := MergeUpgradeCoverageMatrices(root, time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NoError(t, merged.Validate())
	require.Len(t, merged.SourceMatrices, 2)
	for _, row := range merged.Rows {
		require.Equal(t, UpgradeCoverageStatusPassed, row.Status, row.Area)
		for _, claim := range row.QueryCoverage {
			for _, artifactPath := range append(
				append([]string(nil), claim.EvidencePaths...),
				claim.HistoricalHeightEvidencePaths...,
			) {
				require.Contains(t, artifactPath, "run-", "%s/%s", row.Area, claim.Boundary)
			}
			for _, reference := range append(
				append([]UpgradeQueryEvidenceReference(nil), claim.Evidence...),
				claim.HistoricalHeightEvidence...,
			) {
				require.Contains(t, reference.ArtifactPath, "run-", "%s/%s/%s", row.Area, claim.Boundary, reference.Step)
			}
		}
		for _, phase := range row.Phases {
			require.Equal(t, UpgradeCoverageStatusPassed, phase.Status, "%s/%s", row.Area, phase.Name)
			require.NotEmpty(t, phase.ArtifactPaths)
			for _, artifactPath := range phase.ArtifactPaths {
				require.Contains(t, artifactPath, "run-")
			}
		}
	}
}

func TestMergeUpgradeCoverageMatricesRejectsMissingQueryEvidence(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	matrix := validUpgradeCoverageMatrix()
	writeCoverageFixture(t, root, "run-connected", matrix)
	queryEvidence := matrix.Rows[0].QueryCoverage[0].EvidencePaths[0]
	require.NoError(t, os.Remove(filepath.Join(root, "run-connected", filepath.FromSlash(queryEvidence))))

	_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
	require.ErrorContains(t, err, "query boundary")
	require.ErrorContains(t, err, "claimed artifact")
}

func TestMergeUpgradeCoverageMatricesRejectsQueryTransportClaimMismatch(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	matrix := validUpgradeCoverageMatrix()
	writeCoverageFixture(t, root, "run-connected", matrix)
	reference := matrix.Rows[0].QueryCoverage[0].Evidence[0]
	queryPath := filepath.Join(root, "run-connected", filepath.FromSlash(reference.ArtifactPath))
	contents, err := os.ReadFile(queryPath)
	require.NoError(t, err)
	contents = bytes.Replace(contents, []byte(`"boundary":"cli"`), []byte(`"boundary":"rest"`), 1)
	require.NoError(t, os.WriteFile(queryPath, contents, 0o600))

	_, err = MergeUpgradeCoverageMatrices(root, time.Now().UTC())
	require.ErrorContains(t, err, "structured query artifact")
	require.ErrorContains(t, err, "has no boundary=CLI")
}

func TestValidateUpgradeCoverageQueryRecordFailsClosed(t *testing.T) {
	t.Parallel()
	reference := UpgradeQueryEvidenceReference{
		ArtifactPath:     "queries/results.jsonl",
		Boundary:         UpgradeQueryBoundaryCLI,
		Step:             "height-pinned-cli",
		HistoricalHeight: true,
	}
	valid := upgradeCoverageQueryRecord{
		RecordedAt:       time.Now().UTC(),
		Boundary:         "cli",
		Step:             reference.Step,
		Height:           77,
		HistoricalHeight: true,
		Request:          json.RawMessage(`{"arguments":["staking","pool","--height","77"]}`),
		Response:         json.RawMessage(`{"pool":{}}`),
	}
	require.NoError(t, validateUpgradeCoverageQueryRecord(valid, reference, false))
	require.ErrorContains(t, validateUpgradeCoverageQueryRecord(valid, reference, true), "server-validated height")
	restReference := UpgradeQueryEvidenceReference{
		ArtifactPath:     "queries/results.jsonl",
		Boundary:         UpgradeQueryBoundaryREST,
		Step:             "height-pinned-rest",
		HistoricalHeight: true,
	}
	restRecord := upgradeCoverageQueryRecord{
		RecordedAt:       time.Now().UTC(),
		Boundary:         "rest",
		Step:             restReference.Step,
		Height:           77,
		HistoricalHeight: true,
		Request:          json.RawMessage(`{"method":"GET","path":"/fixture","height":77}`),
		Response:         json.RawMessage(`{"ok":true}`),
		Status:           200,
		Metadata:         json.RawMessage(`{"grpc_block_height":"77"}`),
	}
	require.NoError(t, validateUpgradeCoverageQueryRecord(restRecord, restReference, true))

	for _, test := range []struct {
		name   string
		mutate func(*upgradeCoverageQueryRecord)
		want   string
	}{
		{name: "request", mutate: func(record *upgradeCoverageQueryRecord) { record.Request = nil }, want: "request payload"},
		{name: "response", mutate: func(record *upgradeCoverageQueryRecord) { record.Response = nil }, want: "response payload"},
		{name: "recorded error", mutate: func(record *upgradeCoverageQueryRecord) { record.Error = "transport failed" }, want: "recorded error"},
		{name: "height flag", mutate: func(record *upgradeCoverageQueryRecord) {
			record.Request = json.RawMessage(`{"arguments":["staking","pool","--height","78"]}`)
		}, want: "request height=78"},
		{name: "historical marker", mutate: func(record *upgradeCoverageQueryRecord) { record.HistoricalHeight = false }, want: "disagrees with height"},
	} {
		t.Run(test.name, func(t *testing.T) {
			record := valid
			test.mutate(&record)
			err := validateUpgradeCoverageQueryRecord(record, reference, false)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestMergeUpgradeCoverageMatricesRejectsMissingClaimedArtifact(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	matrix := validUpgradeCoverageMatrix()
	writeCoverageFixture(t, root, "run-connected", matrix)
	missing := filepath.Join(root, "run-connected", filepath.FromSlash(matrix.Rows[0].Phases[0].ArtifactPaths[0]))
	require.NoError(t, os.Remove(missing))

	_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
	require.ErrorContains(t, err, "claimed artifact")
}

func TestMergeUpgradeCoverageMatricesRejectsCheckpointWithoutMeaningfulObservation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		contents []byte
		want     string
	}{
		{name: "missing", contents: []byte(`{"height":77}`), want: "no common observation"},
		{name: "height mismatch", contents: []byte(`{
  "recorded_at":"2026-08-04T12:00:00Z",
  "height":78,
  "observation":{
    "observed_at":"2026-08-04T12:00:00Z",
    "node":"fullnode-0",
    "query_boundary":"cometbft-rpc",
    "height":77,
    "block_id":"AABB",
    "app_hash":"CCDD"
  }
}`), want: "does not match height"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			matrix := validUpgradeCoverageMatrix()
			writeCoverageFixture(t, root, "run-connected", matrix)
			checkpointPath := matrix.Rows[0].Phases[0].ArtifactPaths[0]
			require.NoError(t, os.WriteFile(
				filepath.Join(root, "run-connected", filepath.FromSlash(checkpointPath)),
				test.contents,
				0o600,
			))

			_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestMergeUpgradeCoverageMatricesValidatesDIDAndEmptyStoreCheckpoints(t *testing.T) {
	t.Parallel()
	for _, artifactPath := range []string{
		"upgrade/did/pre-upgrade.json",
		"upgrade/did/post-upgrade.json",
		"upgrade/did/post-restart.json",
		"upgrade/legacy-pnft-normal-empty.json",
		"upgrade/nft-empty-post-upgrade-preservation.json",
	} {
		artifactPath := artifactPath
		t.Run(artifactPath, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			matrix := validUpgradeCoverageMatrix()
			matrix.Rows[0].Phases[0].ArtifactPaths = []string{artifactPath}
			writeCoverageFixture(t, root, "run-connected", matrix)
			recordedPath := filepath.Join(root, "run-connected", filepath.FromSlash(artifactPath))
			require.NoError(t, os.WriteFile(recordedPath, []byte(`{"semantic_state":"present"}`), 0o600))

			_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
			require.ErrorContains(t, err, "checkpoint artifact has no common observation")
		})
	}
}

func TestMergeUpgradeCoverageMatricesRejectsUnclosedCoverageGap(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	matrix := validUpgradeCoverageMatrix()
	setCoverageRowNotRun(&matrix, UpgradeCoverageAreaIBCTransfer, "IBC lane was not run")
	writeCoverageFixture(t, root, "run-connected", matrix)

	_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
	require.ErrorContains(t, err, "IBC/transfer")
	require.ErrorContains(t, err, "no passed evidence")
}

func TestMergeUpgradeCoverageMatricesRejectsFailedOrUncleanSourceRun(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	matrix := validUpgradeCoverageMatrix()
	writeCoverageFixture(t, root, "run-failed", matrix)
	failedManifest := []byte(`{"state":"failed","failed":true,"cleanup":{"result":"succeeded"}}`)
	require.NoError(t, os.WriteFile(filepath.Join(root, "run-failed", "manifest.json"), failedManifest, 0o600))

	_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
	require.ErrorContains(t, err, "successful cleaned run")
}

func TestMergeUpgradeCoverageMatricesRejectsPhasesSplicedAcrossDifferentObjects(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	first := validUpgradeCoverageMatrix()
	second := validUpgradeCoverageMatrix()
	firstRow := coverageRowByArea(first.Rows, UpgradeCoverageAreaAuthBank)
	secondRow := coverageRowByArea(second.Rows, UpgradeCoverageAreaAuthBank)
	require.NotNil(t, firstRow)
	require.NotNil(t, secondRow)
	firstRow.StateObjectIDs = []string{"account:first"}
	secondRow.StateObjectIDs = []string{"account:second"}
	for _, index := range []int{3, 4} {
		firstRow.Phases[index].Status = UpgradeCoverageStatusNotRun
		firstRow.Phases[index].ArtifactPaths = nil
		firstRow.Phases[index].Reason = "only the first three phases used account:first"
	}
	firstRow.Status = UpgradeCoverageStatusNotRun
	for _, index := range []int{0, 1, 2} {
		secondRow.Phases[index].Status = UpgradeCoverageStatusNotRun
		secondRow.Phases[index].ArtifactPaths = nil
		secondRow.Phases[index].Reason = "only the last two phases used account:second"
	}
	secondRow.Status = UpgradeCoverageStatusNotRun
	writeCoverageFixture(t, root, "run-first", first)
	writeCoverageFixture(t, root, "run-second", second)

	_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
	require.ErrorContains(t, err, "single source row")
}

func TestMergeUpgradeCoverageMatricesRejectsIntermediateSymlinkEscape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	matrix := validUpgradeCoverageMatrix()
	writeCoverageFixture(t, root, "run-symlink", matrix)
	row := coverageRowByArea(matrix.Rows, UpgradeCoverageAreaAuthBank)
	require.NotNil(t, row)
	linkedDirectory := filepath.Join(root, "run-symlink", "state-checkpoints", "auth-bank")
	require.NoError(t, os.RemoveAll(linkedDirectory))
	for _, phase := range row.Phases {
		for _, artifactPath := range phase.ArtifactPaths {
			fileName := filepath.Base(filepath.FromSlash(artifactPath))
			require.NoError(t, os.WriteFile(filepath.Join(outside, fileName), []byte("{}\n"), 0o600))
		}
	}
	if err := os.Symlink(outside, linkedDirectory); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, err := MergeUpgradeCoverageMatrices(root, time.Now().UTC())
	require.ErrorContains(t, err, "resolved run root")
}

func TestWriteMergedUpgradeCoverageMatrixRejectsSymlinkedOutputDirectory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	matrix := validUpgradeCoverageMatrix()
	writeCoverageFixture(t, root, "run-output", matrix)
	if err := os.Symlink(outside, filepath.Join(root, "upgrade")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, err := WriteMergedUpgradeCoverageMatrix(root, UpgradeCoverageMatrixArtifactPath, time.Now().UTC())
	require.ErrorContains(t, err, "symlink")
}

func setCoverageRowNotRun(matrix *UpgradeCoverageMatrix, area UpgradeCoverageArea, reason string) {
	for rowIndex := range matrix.Rows {
		if matrix.Rows[rowIndex].Area != area {
			continue
		}
		matrix.Rows[rowIndex].Status = UpgradeCoverageStatusNotRun
		for phaseIndex := range matrix.Rows[rowIndex].Phases {
			matrix.Rows[rowIndex].Phases[phaseIndex].Status = UpgradeCoverageStatusNotRun
			matrix.Rows[rowIndex].Phases[phaseIndex].ArtifactPaths = nil
			matrix.Rows[rowIndex].Phases[phaseIndex].Reason = reason
		}
	}
}

func writeCoverageFixture(t *testing.T, root, runID string, matrix UpgradeCoverageMatrix) {
	t.Helper()
	runDir := filepath.Join(root, runID)
	plainQueryArtifacts := make(map[string]struct{})
	structuredQueryArtifacts := make(map[string]map[string]UpgradeQueryEvidenceReference)
	for _, row := range matrix.Rows {
		for _, claim := range row.QueryCoverage {
			for _, artifactPath := range append(
				append([]string(nil), claim.EvidencePaths...),
				claim.HistoricalHeightEvidencePaths...,
			) {
				plainQueryArtifacts[artifactPath] = struct{}{}
			}
			for _, reference := range append(
				append([]UpgradeQueryEvidenceReference(nil), claim.Evidence...),
				claim.HistoricalHeightEvidence...,
			) {
				if structuredQueryArtifacts[reference.ArtifactPath] == nil {
					structuredQueryArtifacts[reference.ArtifactPath] = make(map[string]UpgradeQueryEvidenceReference)
				}
				heightKind := "latest"
				if reference.HistoricalHeight {
					heightKind = "historical"
				}
				key := string(reference.Boundary) + "\x00" + reference.Step + "\x00" + heightKind
				structuredQueryArtifacts[reference.ArtifactPath][key] = reference
			}
		}
		for _, phase := range row.Phases {
			if phase.Status != UpgradeCoverageStatusPassed {
				continue
			}
			for _, artifactPath := range phase.ArtifactPaths {
				absolute := filepath.Join(runDir, filepath.FromSlash(artifactPath))
				require.NoError(t, os.MkdirAll(filepath.Dir(absolute), 0o700))
				contents := []byte("{}\n")
				if isUpgradeCheckpointArtifactPath(artifactPath) {
					contents = []byte(`{
  "recorded_at":"2026-08-04T12:00:00Z",
  "height":77,
  "observation":{
    "observed_at":"2026-08-04T12:00:00Z",
    "node":"fullnode-0",
    "query_boundary":"cometbft-rpc",
    "height":77,
    "block_id":"AABB",
    "app_hash":"CCDD"
  }
}
`)
				}
				require.NoError(t, os.WriteFile(absolute, contents, 0o600))
			}
		}
	}
	for artifactPath := range plainQueryArtifacts {
		if _, structured := structuredQueryArtifacts[artifactPath]; structured {
			continue
		}
		absolute := filepath.Join(runDir, filepath.FromSlash(artifactPath))
		require.NoError(t, os.MkdirAll(filepath.Dir(absolute), 0o700))
		require.NoError(t, os.WriteFile(absolute, []byte("{}\n"), 0o600))
	}
	for artifactPath, references := range structuredQueryArtifacts {
		var contents []byte
		for _, reference := range references {
			height := int64(0)
			if reference.HistoricalHeight {
				height = 77
			}
			boundary := strings.ToLower(string(reference.Boundary))
			request := map[string]any{"arguments": []string{"query", "fixture"}}
			status := 0
			metadata := map[string]any{}
			if height > 0 {
				request["arguments"] = []string{"query", "fixture", "--height", "77"}
			}
			if boundary == "rest" {
				request = map[string]any{"method": "GET", "path": "/fixture", "height": height}
				status = 200
				if height > 0 {
					metadata["grpc_block_height"] = "77"
				}
			}
			record, err := json.Marshal(map[string]any{
				"recorded_at":       time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC),
				"boundary":          boundary,
				"step":              reference.Step,
				"height":            height,
				"historical_height": reference.HistoricalHeight,
				"request":           request,
				"response":          map[string]any{"ok": true},
				"status":            status,
				"metadata":          metadata,
			})
			require.NoError(t, err)
			contents = append(contents, record...)
			contents = append(contents, '\n')
		}
		absolute := filepath.Join(runDir, filepath.FromSlash(artifactPath))
		require.NoError(t, os.MkdirAll(filepath.Dir(absolute), 0o700))
		require.NoError(t, os.WriteFile(absolute, contents, 0o600))
	}
	contents, err := json.MarshalIndent(matrix, "", "  ")
	require.NoError(t, err)
	matrixPath := filepath.Join(runDir, filepath.FromSlash(UpgradeCoverageMatrixArtifactPath))
	require.NoError(t, os.MkdirAll(filepath.Dir(matrixPath), 0o700))
	require.NoError(t, os.WriteFile(matrixPath, contents, 0o600))
	manifest := []byte(`{"state":"cleaned","failed":false,"cleanup":{"result":"succeeded"}}`)
	cleanup := []byte(`{"state":"completed","result":"succeeded"}`)
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "manifest.json"), manifest, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "cleanup.json"), cleanup, 0o600))
}
