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

const upgradeCoverageMatrixReadLimit = 4 << 20

type upgradeCoverageMatrixSource struct {
	relativePath string
	runDir       string
	matrix       UpgradeCoverageMatrix
}

// MergeUpgradeCoverageMatrices discovers direct run-* children of root,
// verifies every artifact claimed by a passed phase, and closes each P0/P1
// phase only when at least one isolated lane provides real passed evidence.
func MergeUpgradeCoverageMatrices(root string, recordedAt time.Time) (UpgradeCoverageMatrix, error) {
	if recordedAt.IsZero() {
		return UpgradeCoverageMatrix{}, errors.New("merged coverage recorded_at is required")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return UpgradeCoverageMatrix{}, fmt.Errorf("resolve coverage root: %w", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return UpgradeCoverageMatrix{}, fmt.Errorf("read coverage root %s: %w", root, err)
	}
	var sources []upgradeCoverageMatrixSource
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "run-") {
			continue
		}
		runDir := filepath.Join(root, entry.Name())
		matrixPath := filepath.Join(runDir, filepath.FromSlash(UpgradeCoverageMatrixArtifactPath))
		info, statErr := os.Lstat(matrixPath)
		if errors.Is(statErr, fs.ErrNotExist) {
			continue
		}
		if statErr != nil {
			return UpgradeCoverageMatrix{}, fmt.Errorf("inspect coverage matrix %s: %w", matrixPath, statErr)
		}
		if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > upgradeCoverageMatrixReadLimit {
			return UpgradeCoverageMatrix{}, fmt.Errorf("coverage matrix %s must be a non-empty regular file no larger than %d bytes", matrixPath, upgradeCoverageMatrixReadLimit)
		}
		contents, readErr := os.ReadFile(matrixPath)
		if readErr != nil {
			return UpgradeCoverageMatrix{}, fmt.Errorf("read coverage matrix %s: %w", matrixPath, readErr)
		}
		var matrix UpgradeCoverageMatrix
		if decodeErr := json.Unmarshal(contents, &matrix); decodeErr != nil {
			return UpgradeCoverageMatrix{}, fmt.Errorf("decode coverage matrix %s: %w", matrixPath, decodeErr)
		}
		if validateErr := matrix.Validate(); validateErr != nil {
			return UpgradeCoverageMatrix{}, fmt.Errorf("validate coverage matrix %s: %w", matrixPath, validateErr)
		}
		if len(matrix.SourceMatrices) != 0 {
			return UpgradeCoverageMatrix{}, fmt.Errorf("coverage source %s is already an aggregate", matrixPath)
		}
		if statusErr := validateUpgradeCoverageSourceRun(runDir); statusErr != nil {
			return UpgradeCoverageMatrix{}, fmt.Errorf("coverage source %s is not a successful cleaned run: %w", matrixPath, statusErr)
		}
		sources = append(sources, upgradeCoverageMatrixSource{
			relativePath: filepath.ToSlash(filepath.Join(entry.Name(), filepath.FromSlash(UpgradeCoverageMatrixArtifactPath))),
			runDir:       runDir,
			matrix:       matrix,
		})
	}
	if len(sources) == 0 {
		return UpgradeCoverageMatrix{}, fmt.Errorf("no run-owned %s files found directly below %s", UpgradeCoverageMatrixArtifactPath, root)
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].relativePath < sources[j].relativePath })
	baseline := sources[0].matrix
	for _, source := range sources[1:] {
		if source.matrix.SchemaVersion != baseline.SchemaVersion ||
			source.matrix.UpgradeName != baseline.UpgradeName ||
			source.matrix.SourceVersion != baseline.SourceVersion ||
			source.matrix.TargetVersion != baseline.TargetVersion {
			return UpgradeCoverageMatrix{}, fmt.Errorf(
				"coverage matrix %s metadata does not match %s",
				source.relativePath,
				sources[0].relativePath,
			)
		}
	}

	merged := UpgradeCoverageMatrix{
		SchemaVersion:  baseline.SchemaVersion,
		RecordedAt:     recordedAt.UTC(),
		UpgradeName:    baseline.UpgradeName,
		SourceVersion:  baseline.SourceVersion,
		TargetVersion:  baseline.TargetVersion,
		SourceMatrices: make([]string, 0, len(sources)),
		Rows:           make([]UpgradeCoverageRow, 0, len(upgradeCoverageAreaRequirements)),
	}
	for _, source := range sources {
		merged.SourceMatrices = append(merged.SourceMatrices, source.relativePath)
	}
	for _, requirement := range upgradeCoverageAreaRequirements {
		row, mergeErr := mergeUpgradeCoverageRow(root, requirement, sources)
		if mergeErr != nil {
			return UpgradeCoverageMatrix{}, mergeErr
		}
		merged.Rows = append(merged.Rows, row)
	}
	if err := merged.Validate(); err != nil {
		return UpgradeCoverageMatrix{}, fmt.Errorf("validate merged coverage matrix: %w", err)
	}
	return merged, nil
}

func validateUpgradeCoverageSourceRun(runDir string) error {
	type manifestCleanup struct {
		Result string `json:"result"`
	}
	type runManifest struct {
		State   string          `json:"state"`
		Failed  *bool           `json:"failed"`
		Cleanup manifestCleanup `json:"cleanup"`
	}
	type runCleanup struct {
		State  string `json:"state"`
		Result string `json:"result"`
	}
	var manifest runManifest
	if err := readUpgradeCoverageRunJSON(filepath.Join(runDir, "manifest.json"), &manifest); err != nil {
		return err
	}
	if manifest.State != "cleaned" || manifest.Failed == nil || *manifest.Failed || manifest.Cleanup.Result != "succeeded" {
		return fmt.Errorf("manifest state=%q failed=%v cleanup=%q", manifest.State, manifest.Failed, manifest.Cleanup.Result)
	}
	var cleanup runCleanup
	if err := readUpgradeCoverageRunJSON(filepath.Join(runDir, "cleanup.json"), &cleanup); err != nil {
		return err
	}
	if cleanup.State != "completed" || cleanup.Result != "succeeded" {
		return fmt.Errorf("cleanup state=%q result=%q", cleanup.State, cleanup.Result)
	}
	return nil
}

func readUpgradeCoverageRunJSON(filePath string, target any) error {
	info, err := os.Lstat(filePath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > upgradeCoverageMatrixReadLimit {
		return fmt.Errorf("%s must be a non-empty regular bounded file", filePath)
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, target); err != nil {
		return fmt.Errorf("decode %s: %w", filePath, err)
	}
	return nil
}

func mergeUpgradeCoverageRow(
	root string,
	requirement upgradeCoverageAreaRequirement,
	sources []upgradeCoverageMatrixSource,
) (UpgradeCoverageRow, error) {
	selectedSourceIndex := -1
	var selectedRow *UpgradeCoverageRow
	var observations []string
	for sourceIndex := range sources {
		row := coverageRowByArea(sources[sourceIndex].matrix.Rows, requirement.Area)
		if row == nil {
			return UpgradeCoverageRow{}, fmt.Errorf("coverage matrix %s is missing area %q", sources[sourceIndex].relativePath, requirement.Area)
		}
		observations = append(observations, fmt.Sprintf("%s=%s", sources[sourceIndex].relativePath, row.Status))
		if selectedRow == nil && row.Status == UpgradeCoverageStatusPassed {
			selectedSourceIndex = sourceIndex
			selectedRow = row
		}
	}
	if selectedRow == nil {
		return UpgradeCoverageRow{}, fmt.Errorf(
			"coverage area %q has no passed evidence from a single source row (%s)",
			requirement.Area, strings.Join(observations, ", "),
		)
	}
	selectedSource := sources[selectedSourceIndex]
	queryCoverage := make([]UpgradeQueryCoverage, 0, len(selectedRow.QueryCoverage))
	for _, claim := range selectedRow.QueryCoverage {
		mergedClaim := claim
		mergedClaim.EvidencePaths = nil
		for _, artifactPath := range claim.EvidencePaths {
			aggregatePath, validateErr := validateUpgradeCoverageClaim(root, selectedSource.runDir, artifactPath)
			if validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q query boundary %q in %s: %w",
					requirement.Area, claim.Boundary, selectedSource.relativePath, validateErr,
				)
			}
			mergedClaim.EvidencePaths = append(mergedClaim.EvidencePaths, aggregatePath)
		}
		mergedClaim.Evidence = nil
		for _, reference := range claim.Evidence {
			aggregatePath, validateErr := validateUpgradeCoverageClaim(root, selectedSource.runDir, reference.ArtifactPath)
			if validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q query evidence %q in %s: %w",
					requirement.Area, reference.Step, selectedSource.relativePath, validateErr,
				)
			}
			if validateErr := validateUpgradeCoverageQueryEvidence(selectedSource.runDir, reference, false); validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q query evidence %q in %s: %w",
					requirement.Area, reference.Step, selectedSource.relativePath, validateErr,
				)
			}
			mergedReference := reference
			mergedReference.ArtifactPath = aggregatePath
			mergedClaim.Evidence = append(mergedClaim.Evidence, mergedReference)
		}
		mergedClaim.HistoricalHeightEvidencePaths = nil
		for _, artifactPath := range claim.HistoricalHeightEvidencePaths {
			aggregatePath, validateErr := validateUpgradeCoverageClaim(root, selectedSource.runDir, artifactPath)
			if validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q historical query boundary %q in %s: %w",
					requirement.Area, claim.Boundary, selectedSource.relativePath, validateErr,
				)
			}
			mergedClaim.HistoricalHeightEvidencePaths = append(mergedClaim.HistoricalHeightEvidencePaths, aggregatePath)
		}
		mergedClaim.HistoricalHeightEvidence = nil
		for _, reference := range claim.HistoricalHeightEvidence {
			aggregatePath, validateErr := validateUpgradeCoverageClaim(root, selectedSource.runDir, reference.ArtifactPath)
			if validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q historical query evidence %q in %s: %w",
					requirement.Area, reference.Step, selectedSource.relativePath, validateErr,
				)
			}
			if validateErr := validateUpgradeCoverageQueryEvidence(selectedSource.runDir, reference, true); validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q historical query evidence %q in %s: %w",
					requirement.Area, reference.Step, selectedSource.relativePath, validateErr,
				)
			}
			mergedReference := reference
			mergedReference.ArtifactPath = aggregatePath
			mergedClaim.HistoricalHeightEvidence = append(mergedClaim.HistoricalHeightEvidence, mergedReference)
		}
		queryCoverage = append(queryCoverage, mergedClaim)
	}
	phases := make([]UpgradeCoveragePhase, 0, len(requiredUpgradeCoveragePhases))
	for phaseIndex, phaseName := range requiredUpgradeCoveragePhases {
		artifactPaths := make(map[string]struct{})
		phase := selectedRow.Phases[phaseIndex]
		if phase.Status != UpgradeCoverageStatusPassed {
			return UpgradeCoverageRow{}, fmt.Errorf(
				"coverage area %q selected source %s phase %q is %s, want passed",
				requirement.Area, selectedSource.relativePath, phaseName, phase.Status,
			)
		}
		for _, artifactPath := range phase.ArtifactPaths {
			aggregatePath, validateErr := validateUpgradeCoverageClaim(root, selectedSource.runDir, artifactPath)
			if validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q phase %q in %s: %w",
					requirement.Area, phaseName, selectedSource.relativePath, validateErr,
				)
			}
			if validateErr := validateUpgradeCheckpointArtifact(selectedSource.runDir, artifactPath); validateErr != nil {
				return UpgradeCoverageRow{}, fmt.Errorf(
					"coverage area %q phase %q checkpoint %q in %s: %w",
					requirement.Area, phaseName, artifactPath, selectedSource.relativePath, validateErr,
				)
			}
			artifactPaths[aggregatePath] = struct{}{}
		}
		phases = append(phases, UpgradeCoveragePhase{
			Name:          phaseName,
			Status:        UpgradeCoverageStatusPassed,
			ArtifactPaths: sortedCoverageKeys(artifactPaths),
		})
	}
	return UpgradeCoverageRow{
		Area:            requirement.Area,
		Priority:        requirement.Priority,
		Status:          UpgradeCoverageStatusPassed,
		StateObjectIDs:  append([]string(nil), selectedRow.StateObjectIDs...),
		LowerLevelTests: append([]string(nil), selectedRow.LowerLevelTests...),
		QueryCoverage:   queryCoverage,
		Phases:          phases,
	}, nil
}

func validateUpgradeCheckpointArtifact(runDir, artifactPath string) error {
	if !isUpgradeCheckpointArtifactPath(artifactPath) {
		return nil
	}
	var document any
	if err := readUpgradeCoverageRunJSON(filepath.Join(runDir, filepath.FromSlash(artifactPath)), &document); err != nil {
		return fmt.Errorf("read checkpoint artifact: %w", err)
	}
	count, err := validateUpgradeCheckpointObservations(document, "$")
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("checkpoint artifact has no common observation")
	}
	return nil
}

func isUpgradeCheckpointArtifactPath(artifactPath string) bool {
	if strings.HasPrefix(artifactPath, "state-checkpoints/") || strings.Contains(artifactPath, "/checkpoints/") {
		return true
	}
	switch artifactPath {
	case AOLUpgradePreCheckpointArtifactPath,
		AOLUpgradePostPreservationArtifactPath,
		AOLUpgradePostMutationArtifactPath,
		AOLUpgradePostRestartArtifactPath,
		"upgrade/did/pre-upgrade.json",
		"upgrade/did/post-upgrade.json",
		"upgrade/did/post-restart.json",
		"upgrade/legacy-pnft-normal-empty.json",
		"upgrade/nft-empty-post-upgrade-preservation.json",
		"upgrade/staking/preparation.json",
		"upgrade/staking/mutation.json",
		"upgrade/staking/validator-liveness.json",
		"upgrade/staking/post-restart-validation.json",
		"upgrade/authz-feegrant/preparation.json",
		"upgrade/authz-feegrant/mutation.json",
		"upgrade/group-vesting/preparation.json",
		"upgrade/group-vesting/mutation.json",
		"upgrade/group-vesting/post-restart.json",
		"upgrade/system-modules/mutation.json":
		return true
	default:
		return false
	}
}

func validateUpgradeCheckpointObservations(value any, location string) (int, error) {
	switch typed := value.(type) {
	case map[string]any:
		count := 0
		if raw, present := typed["observation"]; present {
			contents, err := json.Marshal(raw)
			if err != nil {
				return 0, fmt.Errorf("%s observation encoding: %w", location, err)
			}
			var observation UpgradeCheckpointObservation
			if err := json.Unmarshal(contents, &observation); err != nil {
				return 0, fmt.Errorf("%s observation decoding: %w", location, err)
			}
			if err := observation.Validate(); err != nil {
				return 0, fmt.Errorf("%s observation: %w", location, err)
			}
			for _, heightField := range []string{"height", "query_height"} {
				if outerHeight, ok := jsonInteger(typed[heightField]); ok && outerHeight != observation.Height {
					return 0, fmt.Errorf(
						"%s observation height %d does not match %s %d",
						location, observation.Height, heightField, outerHeight,
					)
				}
			}
			if rawRecordedAt, ok := typed["recorded_at"].(string); ok && strings.TrimSpace(rawRecordedAt) != "" {
				recordedAt, err := time.Parse(time.RFC3339Nano, rawRecordedAt)
				if err != nil {
					return 0, fmt.Errorf("%s recorded_at is invalid: %w", location, err)
				}
				if !recordedAt.Equal(observation.ObservedAt) {
					return 0, fmt.Errorf("%s observation observed_at does not match recorded_at", location)
				}
			}
			count++
		}
		for key, child := range typed {
			if key == "observation" {
				continue
			}
			childCount, err := validateUpgradeCheckpointObservations(child, location+"."+key)
			if err != nil {
				return 0, err
			}
			count += childCount
		}
		return count, nil
	case []any:
		count := 0
		for index, child := range typed {
			childCount, err := validateUpgradeCheckpointObservations(child, fmt.Sprintf("%s[%d]", location, index))
			if err != nil {
				return 0, err
			}
			count += childCount
		}
		return count, nil
	default:
		return 0, nil
	}
}

func jsonInteger(value any) (int64, bool) {
	number, ok := value.(float64)
	if !ok || number != float64(int64(number)) {
		return 0, false
	}
	return int64(number), true
}

const (
	upgradeCoverageQueryLineLimit   = 32 << 20
	upgradeCoverageQueryRecordLimit = 200_000
)

type upgradeCoverageQueryRecord struct {
	RecordedAt       time.Time       `json:"recorded_at"`
	Boundary         string          `json:"boundary"`
	Step             string          `json:"step"`
	Height           int64           `json:"height"`
	HistoricalHeight bool            `json:"historical_height"`
	Request          json.RawMessage `json:"request"`
	Response         json.RawMessage `json:"response"`
	Status           int             `json:"status"`
	Metadata         json.RawMessage `json:"metadata"`
	Error            string          `json:"error"`
}

func validateUpgradeCoverageQueryEvidence(
	runDir string,
	reference UpgradeQueryEvidenceReference,
	requireServerValidatedHeight bool,
) error {
	absolute := filepath.Join(runDir, filepath.FromSlash(reference.ArtifactPath))
	file, err := os.Open(absolute)
	if err != nil {
		return fmt.Errorf("open structured query artifact %q: %w", reference.ArtifactPath, err)
	}
	defer file.Close()

	wantBoundary := strings.ToLower(string(reference.Boundary))
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64<<10), upgradeCoverageQueryLineLimit)
	recordCount := 0
	matchCount := 0
	for scanner.Scan() {
		recordCount++
		if recordCount > upgradeCoverageQueryRecordLimit {
			return fmt.Errorf("structured query artifact %q exceeds %d records", reference.ArtifactPath, upgradeCoverageQueryRecordLimit)
		}
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			return fmt.Errorf("structured query artifact %q contains a blank record", reference.ArtifactPath)
		}
		var record upgradeCoverageQueryRecord
		if err := json.Unmarshal(line, &record); err != nil {
			return fmt.Errorf("decode structured query artifact %q record %d: %w", reference.ArtifactPath, recordCount, err)
		}
		if record.Boundary != wantBoundary || record.Step != reference.Step {
			continue
		}
		matchCount++
		if matchCount > 1 {
			return fmt.Errorf("structured query evidence boundary=%s step=%q is duplicated", reference.Boundary, reference.Step)
		}
		if err := validateUpgradeCoverageQueryRecord(record, reference, requireServerValidatedHeight); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan structured query artifact %q: %w", reference.ArtifactPath, err)
	}
	if matchCount == 0 {
		return fmt.Errorf(
			"structured query artifact %q has no boundary=%s step=%q record",
			reference.ArtifactPath,
			reference.Boundary,
			reference.Step,
		)
	}
	return nil
}

func validateUpgradeCoverageQueryRecord(
	record upgradeCoverageQueryRecord,
	reference UpgradeQueryEvidenceReference,
	requireServerValidatedHeight bool,
) error {
	if record.RecordedAt.IsZero() {
		return fmt.Errorf("structured query evidence step %q has no recorded_at", reference.Step)
	}
	if strings.TrimSpace(record.Error) != "" {
		return fmt.Errorf("structured query evidence step %q recorded error %q", reference.Step, record.Error)
	}
	if !upgradeCoverageQueryPayloadPresent(record.Request) {
		return fmt.Errorf("structured query evidence step %q has no request payload", reference.Step)
	}
	if !upgradeCoverageQueryPayloadPresent(record.Response) {
		return fmt.Errorf("structured query evidence step %q has no response payload", reference.Step)
	}
	if record.HistoricalHeight != (record.Height > 0) {
		return fmt.Errorf(
			"structured query evidence step %q historical_height=%t disagrees with height=%d",
			reference.Step,
			record.HistoricalHeight,
			record.Height,
		)
	}
	if record.HistoricalHeight != reference.HistoricalHeight {
		return fmt.Errorf(
			"structured query evidence step %q historical_height=%t, claim requires %t",
			reference.Step,
			record.HistoricalHeight,
			reference.HistoricalHeight,
		)
	}
	if err := validateUpgradeCoverageQueryTransportHeight(record, requireServerValidatedHeight); err != nil {
		return fmt.Errorf("structured query evidence step %q: %w", reference.Step, err)
	}
	return nil
}

func upgradeCoverageQueryPayloadPresent(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || !json.Valid(raw) {
		return false
	}
	var value any
	if json.Unmarshal(raw, &value) != nil || value == nil {
		return false
	}
	_, isString := value.(string)
	return !isString
}

func validateUpgradeCoverageQueryTransportHeight(record upgradeCoverageQueryRecord, requireServerValidatedHeight bool) error {
	switch record.Boundary {
	case "cli", "grpc":
		var request struct {
			Arguments []string `json:"arguments"`
		}
		if err := json.Unmarshal(record.Request, &request); err != nil {
			return fmt.Errorf("decode command request: %w", err)
		}
		if len(request.Arguments) > 0 {
			if height := queryCommandHeight(request.Arguments); height != record.Height {
				return fmt.Errorf("command request height=%d, record height=%d", height, record.Height)
			}
			if requireServerValidatedHeight && record.Height > 0 {
				return fmt.Errorf("%s historical response does not expose a server-validated height", record.Boundary)
			}
			return nil
		}
		var metadata struct {
			RequestHeight int64 `json:"request_height"`
		}
		if err := json.Unmarshal(record.Metadata, &metadata); err != nil {
			return fmt.Errorf("decode gRPC request-height metadata: %w", err)
		}
		if metadata.RequestHeight != record.Height {
			return fmt.Errorf("gRPC request height=%d, record height=%d", metadata.RequestHeight, record.Height)
		}
		if requireServerValidatedHeight && record.Height > 0 {
			return fmt.Errorf("gRPC historical response does not expose a server-validated height")
		}
		return nil
	case "rest":
		var request struct {
			Method string `json:"method"`
			Path   string `json:"path"`
			Height int64  `json:"height"`
		}
		if err := json.Unmarshal(record.Request, &request); err != nil {
			return fmt.Errorf("decode REST request: %w", err)
		}
		if request.Method != "GET" || !strings.HasPrefix(request.Path, "/") {
			return fmt.Errorf("REST request method/path are incomplete")
		}
		if request.Height != record.Height {
			return fmt.Errorf("REST request height=%d, record height=%d", request.Height, record.Height)
		}
		if record.Status < 200 || record.Status >= 300 {
			return fmt.Errorf("REST response status=%d, want 2xx", record.Status)
		}
		if record.Height > 0 {
			var metadata struct {
				BlockHeight string `json:"grpc_block_height"`
			}
			if err := json.Unmarshal(record.Metadata, &metadata); err != nil {
				return fmt.Errorf("decode REST response-height metadata: %w", err)
			}
			if metadata.BlockHeight != fmt.Sprint(record.Height) {
				return fmt.Errorf("REST response height=%q, want %d", metadata.BlockHeight, record.Height)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown query boundary %q", record.Boundary)
	}
}

func validateUpgradeCoverageClaim(root, runDir, artifactPath string) (string, error) {
	if err := validateCoverageArtifactPath(artifactPath); err != nil {
		return "", err
	}
	absolute := filepath.Join(runDir, filepath.FromSlash(artifactPath))
	relativeToRun, err := filepath.Rel(runDir, absolute)
	if err != nil || relativeToRun == ".." || strings.HasPrefix(relativeToRun, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("claimed artifact escapes its run root: %q", artifactPath)
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return "", fmt.Errorf("claimed artifact %q is unavailable: %w", artifactPath, err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return "", fmt.Errorf("claimed artifact %q must be a non-empty regular file", artifactPath)
	}
	resolvedRunDir, err := filepath.EvalSymlinks(runDir)
	if err != nil {
		return "", fmt.Errorf("resolve coverage run root: %w", err)
	}
	resolvedArtifact, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", fmt.Errorf("resolve claimed artifact %q: %w", artifactPath, err)
	}
	resolvedRelative, err := filepath.Rel(resolvedRunDir, resolvedArtifact)
	if err != nil || resolvedRelative == ".." || strings.HasPrefix(resolvedRelative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("claimed artifact %q escapes the resolved run root", artifactPath)
	}
	relativeToRoot, err := filepath.Rel(root, absolute)
	if err != nil || relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("claimed artifact %q escapes aggregate root", artifactPath)
	}
	aggregatePath := filepath.ToSlash(relativeToRoot)
	if err := validateCoverageArtifactPath(aggregatePath); err != nil {
		return "", err
	}
	return aggregatePath, nil
}

func coverageRowByArea(rows []UpgradeCoverageRow, area UpgradeCoverageArea) *UpgradeCoverageRow {
	for index := range rows {
		if rows[index].Area == area {
			return &rows[index]
		}
	}
	return nil
}

func sortedCoverageKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

// WriteMergedUpgradeCoverageMatrix writes the aggregate atomically below root.
func WriteMergedUpgradeCoverageMatrix(root, outputPath string, recordedAt time.Time) (UpgradeCoverageMatrix, error) {
	if err := validateCoverageArtifactPath(outputPath); err != nil {
		return UpgradeCoverageMatrix{}, fmt.Errorf("validate aggregate output path: %w", err)
	}
	matrix, err := MergeUpgradeCoverageMatrices(root, recordedAt)
	if err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	output := filepath.Join(root, filepath.FromSlash(outputPath))
	relative, err := filepath.Rel(root, output)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return UpgradeCoverageMatrix{}, errors.New("aggregate output escapes coverage root")
	}
	if err := rejectUpgradeCoverageSymlinkComponents(root, filepath.Dir(output)); err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o700); err != nil {
		return UpgradeCoverageMatrix{}, fmt.Errorf("create aggregate output directory: %w", err)
	}
	if err := rejectUpgradeCoverageSymlinkComponents(root, filepath.Dir(output)); err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	contents, err := json.MarshalIndent(matrix, "", "  ")
	if err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	contents = append(contents, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(output), ".coverage-matrix-*.tmp")
	if err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return UpgradeCoverageMatrix{}, err
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return UpgradeCoverageMatrix{}, err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return UpgradeCoverageMatrix{}, err
	}
	if err := temporary.Close(); err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	if err := os.Rename(temporaryName, output); err != nil {
		return UpgradeCoverageMatrix{}, err
	}
	return matrix, nil
}

func rejectUpgradeCoverageSymlinkComponents(root, targetDirectory string) error {
	relative, err := filepath.Rel(root, targetDirectory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("aggregate output directory escapes coverage root")
	}
	current := root
	if relative == "." {
		return nil
	}
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, fs.ErrNotExist) {
			return nil
		}
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("aggregate output directory contains symlink component %s", current)
		}
		if !info.IsDir() {
			return fmt.Errorf("aggregate output component %s is not a directory", current)
		}
	}
	return nil
}
