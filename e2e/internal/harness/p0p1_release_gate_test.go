package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testP0P1CurrentImageID        = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testP0P1OldImageID            = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	testP0P1ReleaseCurrentImageID = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	testP0P1ReleaseOldImageID     = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	testP0P1DifferentImageID      = "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	testP0P1CurrentBinarySHA256   = "1111111111111111111111111111111111111111111111111111111111111111"
	testP0P1OldBinarySHA256       = "2222222222222222222222222222222222222222222222222222222222222222"
)

func TestP0P1ReleaseGateManifestRequiresEverySuiteAndSharedImmutableImages(t *testing.T) {
	suites := make([]P0P1ReleaseGateSuiteEvidence, 0, len(requiredP0P1ReleaseGateTests))
	for index, testName := range requiredP0P1ReleaseGateTests {
		suite := P0P1ReleaseGateSuiteEvidence{
			TestName:        testName,
			RunPath:         fmt.Sprintf("run-%012x", index+1),
			InitialImageRef: "panacea-e2e-current:gate",
			FinalImageIDs:   []string{testP0P1CurrentImageID},
			NodeCount:       1,
		}
		if _, startsOld := oldImageP0P1ReleaseGateTests[testName]; startsOld {
			suite.InitialImageRef = "panacea-e2e-v2.2.1:gate"
			suite.OldImageIDs = []string{testP0P1OldImageID}
			suite.SwitchCount = 1
			suite.SwitchedNodes = []string{"node-0"}
		}
		suites = append(suites, suite)
	}
	valid := P0P1ReleaseGateManifest{
		SchemaVersion:          P0P1ReleaseGateSchemaVersion,
		RecordedAt:             time.Now().UTC(),
		SourceCommit:           "0123456789abcdef0123456789abcdef01234567",
		SourceClean:            true,
		CoverageMatrix:         UpgradeCoverageMatrixArtifactPath,
		ReleaseBuildManifest:   "release-20260804000000-1/release-hardening-manifest.json",
		ReleaseHostPlatform:    "linux/amd64",
		ReleaseHostIdentity:    "release-20260804000000-1/host-image-identity.json",
		ReleaseHostImages:      testP0P1ReleaseHostImages(),
		ReleaseImages:          testP0P1ReleasePlatformImages("0123456789abcdef0123456789abcdef01234567"),
		CurrentImageID:         testP0P1CurrentImageID,
		OldImageID:             testP0P1OldImageID,
		CurrentInitialImageRef: "panacea-e2e-current:gate",
		OldInitialImageRef:     "panacea-e2e-v2.2.1:gate",
		RequiredSuites:         suites,
	}
	require.NoError(t, valid.Validate())

	missing := valid
	missing.RequiredSuites = append([]P0P1ReleaseGateSuiteEvidence(nil), suites[:len(suites)-1]...)
	require.ErrorContains(t, missing.Validate(), "want")

	mixedCurrent := valid
	mixedCurrent.RequiredSuites = append([]P0P1ReleaseGateSuiteEvidence(nil), suites...)
	mixedCurrent.RequiredSuites[0].FinalImageIDs = []string{testP0P1OldImageID}
	require.ErrorContains(t, mixedCurrent.Validate(), "final image IDs")

	missingSwitch := valid
	missingSwitch.RequiredSuites = append([]P0P1ReleaseGateSuiteEvidence(nil), suites...)
	missingSwitch.RequiredSuites[0].SwitchCount = 0
	require.ErrorContains(t, missingSwitch.Validate(), "old-to-current")
}

func testP0P1ReleaseHostImages() []ReleaseHostImageIdentityEntry {
	return []ReleaseHostImageIdentityEntry{
		{
			Kind: "current", FunctionalImageRef: "panacea-e2e-current:gate",
			FunctionalImageID: testP0P1CurrentImageID, FunctionalBinarySHA256: testP0P1CurrentBinarySHA256,
			ReleaseImageRef: "panacea-e2e-release-current-amd64:release-20260804000000-1",
			ReleaseImageID:  testP0P1ReleaseCurrentImageID, ReleaseBinarySHA256: testP0P1CurrentBinarySHA256,
		},
		{
			Kind: "v2.2.1", FunctionalImageRef: "panacea-e2e-v2.2.1:gate",
			FunctionalImageID: testP0P1OldImageID, FunctionalBinarySHA256: testP0P1OldBinarySHA256,
			ReleaseImageRef: "panacea-e2e-release-v2.2.1-amd64:release-20260804000000-1",
			ReleaseImageID:  testP0P1ReleaseOldImageID, ReleaseBinarySHA256: testP0P1OldBinarySHA256,
		},
	}
}

func testP0P1ReleasePlatformImages(sourceCommit string) []ReleasePlatformImageEvidence {
	const oldCommit = "89abcdef0123456789abcdef0123456789abcdef"
	var images []ReleasePlatformImageEvidence
	for _, platform := range []string{"linux/amd64", "linux/arm64"} {
		suffix := platform[len("linux/"):]
		for _, kind := range []string{"current", "v2.2.1"} {
			imageID := testP0P1ReleaseCurrentImageID
			binarySHA256 := testP0P1CurrentBinarySHA256
			version := "2.3.0"
			commit := sourceCommit
			if kind == "v2.2.1" {
				imageID = testP0P1ReleaseOldImageID
				binarySHA256 = testP0P1OldBinarySHA256
				version = "2.2.1"
				commit = oldCommit
			}
			images = append(images, ReleasePlatformImageEvidence{
				Kind: kind, Platform: platform,
				ImageRef:    "panacea-e2e-release-" + kind + "-" + suffix + ":release-20260804000000-1",
				ImageDigest: imageID, ImageID: imageID, BinarySHA256: binarySHA256,
				Version: version, SourceCommit: commit,
			})
		}
	}
	return images
}

func TestDiscoverP0P1ReleaseGateSuitesRejectsMissingDuplicateAndMixedRuns(t *testing.T) {
	root := t.TempDir()
	for index, testName := range requiredP0P1ReleaseGateTests {
		writeP0P1GateRunFixture(t, root, fmt.Sprintf("run-%012x", index+1), testName, testP0P1CurrentImageID)
	}
	suites, err := discoverP0P1ReleaseGateSuites(root)
	require.NoError(t, err)
	require.Len(t, suites, len(requiredP0P1ReleaseGateTests))
	for _, suite := range suites {
		require.Equal(t, []string{testP0P1CurrentImageID}, suite.FinalImageIDs)
		if _, startsOld := oldImageP0P1ReleaseGateTests[suite.TestName]; startsOld {
			require.Equal(t, []string{testP0P1OldImageID}, suite.OldImageIDs)
			require.Equal(t, suite.NodeCount, suite.SwitchCount)
			require.Len(t, suite.SwitchedNodes, suite.NodeCount)
		}
	}

	writeP0P1GateRunFixture(t, root, "run-ffffffffffff", requiredP0P1ReleaseGateTests[0], testP0P1CurrentImageID)
	_, err = discoverP0P1ReleaseGateSuites(root)
	require.ErrorContains(t, err, "duplicate")
}

func TestDiscoverP0P1ReleaseGateSuitesAllowsKnownFunctionalSupplementAndRejectsUnknown(t *testing.T) {
	root := t.TempDir()
	for index, testName := range requiredP0P1ReleaseGateTests {
		writeP0P1GateRunFixture(t, root, fmt.Sprintf("run-%012x", index+1), testName, testP0P1CurrentImageID)
	}
	for index, testName := range supplementalP0P1ReleaseGateTests {
		writeP0P1GateRunFixture(t, root, fmt.Sprintf("run-%012x", index+100), testName, testP0P1CurrentImageID)
	}

	suites, err := discoverP0P1ReleaseGateSuites(root)
	require.NoError(t, err)
	require.Len(t, suites, len(requiredP0P1ReleaseGateTests))

	writeP0P1GateRunFixture(t, root, "run-bbbbbbbbbbbb", "TestUnknownLiveSuite", testP0P1CurrentImageID)
	_, err = discoverP0P1ReleaseGateSuites(root)
	require.ErrorContains(t, err, `unexpected run "TestUnknownLiveSuite"`)
}

func TestReadP0P1SwitchImageIDsRequiresExactlyOneSwitchPerDeclaredNode(t *testing.T) {
	runDir := t.TempDir()
	finalNodes := map[string]string{
		"node-0": testP0P1CurrentImageID,
		"node-1": testP0P1CurrentImageID,
	}
	writeSwitches := func(nodes ...string) {
		t.Helper()
		var contents []byte
		for _, node := range nodes {
			record, err := json.Marshal(map[string]any{
				"plan": map[string]any{"node": node}, "old_image_id": testP0P1OldImageID,
				"new_image_id": testP0P1CurrentImageID,
			})
			require.NoError(t, err)
			contents = append(contents, append(record, '\n')...)
		}
		writeP0P1GateTestFile(t, filepath.Join(runDir, "upgrade", "node-switches.jsonl"), string(contents))
	}

	writeSwitches("node-0", "node-1")
	summary, err := readP0P1SwitchImageIDs(runDir, finalNodes, true)
	require.NoError(t, err)
	require.Equal(t, []string{"node-0", "node-1"}, summary.Nodes)
	require.Equal(t, 2, summary.Count)

	writeSwitches("node-0")
	_, err = readP0P1SwitchImageIDs(runDir, finalNodes, true)
	require.ErrorContains(t, err, "want exactly 2")

	writeSwitches("node-0", "node-0")
	_, err = readP0P1SwitchImageIDs(runDir, finalNodes, true)
	require.ErrorContains(t, err, "more than once")

	writeSwitches("node-0", "node-unknown")
	_, err = readP0P1SwitchImageIDs(runDir, finalNodes, true)
	require.ErrorContains(t, err, "undeclared node")
}

func TestWriteP0P1ReleaseGateManifestValidatesCoverageReleaseBuildAndLiveRuns(t *testing.T) {
	root := t.TempDir()
	for index, testName := range requiredP0P1ReleaseGateTests {
		writeP0P1GateRunFixture(t, root, fmt.Sprintf("run-%012x", index+1), testName, testP0P1CurrentImageID)
	}
	matrix := validUpgradeCoverageMatrix()
	matrix.SourceMatrices = []string{"run-000000000001/upgrade/coverage-matrix.json"}
	require.NoError(t, matrix.Validate())
	writeP0P1GateTestJSON(t, filepath.Join(root, filepath.FromSlash(UpgradeCoverageMatrixArtifactPath)), matrix)

	const sourceCommit = "0123456789abcdef0123456789abcdef01234567"
	releaseDir := filepath.Join(root, "release-20260804000000-1")
	writeP0P1ReleaseBuildFixture(t, releaseDir, sourceCommit)

	manifest, err := WriteP0P1ReleaseGateManifest(root, sourceCommit, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Equal(t, testP0P1CurrentImageID, manifest.CurrentImageID)
	require.Equal(t, testP0P1OldImageID, manifest.OldImageID)
	require.Len(t, manifest.RequiredSuites, len(requiredP0P1ReleaseGateTests))
	contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(P0P1ReleaseGateArtifactPath)))
	require.NoError(t, err)
	var persisted P0P1ReleaseGateManifest
	require.NoError(t, json.Unmarshal(contents, &persisted))
	require.Equal(t, manifest, persisted)

	releaseIdentity := ReleaseHostImageIdentity{
		SchemaVersion: ReleaseHostImageIdentitySchemaVersion,
		HostPlatform:  "linux/amd64",
		Images:        testP0P1ReleaseHostImages(),
	}
	releaseIdentity.Images[0].FunctionalImageID = testP0P1DifferentImageID
	writeP0P1GateTestJSON(t, filepath.Join(releaseDir, ReleaseHostImageIdentityArtifactPath), releaseIdentity)
	writeP0P1GateTestJSON(t, filepath.Join(releaseDir, "functional-current-amd64-image-inspect.json"), []map[string]any{{
		"Id": testP0P1DifferentImageID, "Os": "linux", "Architecture": "amd64",
	}})
	_, err = WriteP0P1ReleaseGateManifest(root, sourceCommit, time.Now().UTC())
	require.ErrorContains(t, err, "current functional suite image")
}

func TestValidateReleaseHardeningArtifactRejectsDirtySourceEvidence(t *testing.T) {
	dir := t.TempDir()
	const sourceCommit = "0123456789abcdef0123456789abcdef01234567"
	writeP0P1ReleaseBuildFixture(t, dir, sourceCommit)
	manifestPath := filepath.Join(dir, "release-hardening-manifest.json")
	require.NoError(t, ValidateReleaseHardeningArtifact(manifestPath))

	writeP0P1GateTestFile(t, filepath.Join(dir, "source-status.txt"), " M app/app.go\n")
	require.ErrorContains(t, ValidateReleaseHardeningArtifact(manifestPath), "must be an empty regular file")
}

func TestValidateReleaseHardeningArtifactRejectsNonHostProvenanceMismatch(t *testing.T) {
	const sourceCommit = "0123456789abcdef0123456789abcdef01234567"
	t.Run("metadata digest", func(t *testing.T) {
		dir := t.TempDir()
		writeP0P1ReleaseBuildFixture(t, dir, sourceCommit)
		writeP0P1GateTestJSON(t, filepath.Join(dir, "current-arm64-build-metadata.json"), map[string]any{
			"containerimage.digest": testP0P1DifferentImageID,
		})
		require.ErrorContains(t, ValidateReleaseHardeningArtifact(filepath.Join(dir, "release-hardening-manifest.json")), "want manifest")
	})
	t.Run("platform", func(t *testing.T) {
		dir := t.TempDir()
		writeP0P1ReleaseBuildFixture(t, dir, sourceCommit)
		writeP0P1GateTestJSON(t, filepath.Join(dir, "current-arm64-image-inspect.json"), []map[string]any{{
			"Id": testP0P1ReleaseCurrentImageID, "Os": "linux", "Architecture": "amd64",
		}})
		require.ErrorContains(t, ValidateReleaseHardeningArtifact(filepath.Join(dir, "release-hardening-manifest.json")), "platform")
	})
	t.Run("binary version commit", func(t *testing.T) {
		dir := t.TempDir()
		writeP0P1ReleaseBuildFixture(t, dir, sourceCommit)
		writeP0P1GateTestFile(t, filepath.Join(dir, "current-arm64-version.txt"), "version: 2.3.0\ncommit: 89abcdef0123456789abcdef0123456789abcdef\n")
		require.ErrorContains(t, ValidateReleaseHardeningArtifact(filepath.Join(dir, "release-hardening-manifest.json")), "version/commit")
	})
}

func TestWriteP0P1ReleaseGateFailureRequiresRealCauseAndPersistsStage(t *testing.T) {
	root := t.TempDir()
	recordedAt := time.Now().UTC()
	require.Error(t, WriteP0P1ReleaseGateFailure(root, "", fmt.Errorf("boom"), recordedAt))
	require.Error(t, WriteP0P1ReleaseGateFailure(root, "coverage-merge", nil, recordedAt))
	require.NoError(t, WriteP0P1ReleaseGateFailure(root, "coverage-merge", fmt.Errorf("boom"), recordedAt))
	contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(P0P1ReleaseGateFailureArtifactPath)))
	require.NoError(t, err)
	var failure P0P1ReleaseGateFailure
	require.NoError(t, json.Unmarshal(contents, &failure))
	require.Equal(t, "coverage-merge", failure.Stage)
	require.Equal(t, "boom", failure.Error)
	require.True(t, recordedAt.Equal(failure.RecordedAt))
}

func writeP0P1GateRunFixture(t *testing.T, root, runName, testName, finalImageID string) {
	t.Helper()
	runDir := filepath.Join(root, runName)
	require.NoError(t, os.MkdirAll(filepath.Join(runDir, "nodes", "node-0"), 0o700))
	initialImage := "panacea-e2e-current:gate"
	_, startsOld := oldImageP0P1ReleaseGateTests[testName]
	if startsOld {
		initialImage = "panacea-e2e-v2.2.1:gate"
	}
	writeP0P1GateTestJSON(t, filepath.Join(runDir, "manifest.json"), map[string]any{
		"run_id": runName, "test_name": testName, "state": "cleaned", "failed": false,
		"image": initialImage, "num_validators": 1, "num_full_nodes": 0,
		"cleanup": map[string]any{"result": "succeeded"},
	})
	writeP0P1GateTestJSON(t, filepath.Join(runDir, "cleanup.json"), map[string]any{
		"state": "completed", "result": "succeeded",
	})
	writeP0P1GateTestJSON(t, filepath.Join(runDir, "nodes", "node-0", "container-state.json"), map[string]any{
		"image": finalImageID,
	})
	if startsOld {
		require.NoError(t, os.MkdirAll(filepath.Join(runDir, "upgrade"), 0o700))
		record, err := json.Marshal(map[string]any{
			"plan":         map[string]any{"node": "node-0"},
			"old_image_id": testP0P1OldImageID,
			"new_image_id": finalImageID,
		})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(runDir, "upgrade", "node-switches.jsonl"), append(record, '\n'), 0o600))
	}
}

func writeP0P1GateTestJSON(t *testing.T, filePath string, value any) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o700))
	contents, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filePath, append(contents, '\n'), 0o600))
}

func writeP0P1ReleaseBuildFixture(t *testing.T, dir, sourceCommit string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o700))
	manifest := ReleaseHardeningManifest{
		SchemaVersion:                 ReleaseHardeningManifestSchemaVersion,
		RunID:                         "release-20260804000000-1",
		SourceCommit:                  sourceCommit,
		SourceClean:                   true,
		ColdGoCaches:                  true,
		FreshBuildKitBuilder:          true,
		DockerBuildNetwork:            "none",
		Platforms:                     []string{"linux/amd64", "linux/arm64"},
		VersionAndSmoke:               true,
		MultiarchUpgradeCompatibility: true,
		ImageIndex:                    "image-index.txt",
		HostPlatform:                  "linux/amd64",
		HostImageIdentity:             ReleaseHostImageIdentityArtifactPath,
		Images:                        testP0P1ReleasePlatformImages(sourceCommit),
	}
	writeP0P1GateTestJSON(t, filepath.Join(dir, "release-hardening-manifest.json"), manifest)
	writeP0P1GateTestFile(t, filepath.Join(dir, "status.txt"), "result=passed\nstage=complete\n")
	writeP0P1GateTestFile(t, filepath.Join(dir, "source-commit.txt"), sourceCommit+"\n")
	writeP0P1GateTestFile(t, filepath.Join(dir, "source-commit-final.txt"), sourceCommit+"\n")
	for _, name := range []string{"source-status.txt", "source-diff.patch", "source-status-final.txt", "source-diff-final.patch"} {
		writeP0P1GateTestFile(t, filepath.Join(dir, name), "")
	}
	writeP0P1GateTestFile(t, filepath.Join(dir, "base-images.txt"), `FROM golang:1.26.5-trixie@sha256:8229e3b2cf7fc08878a86977547e3119c173681c3cc4a64c38cf0c6fe0b42fa8 AS build-env
FROM debian:trixie-slim@sha256:3a39a0592364683e6bab97937b72cad5a8fa6dcbbee90edb3bb48c7f8e94f258
`)
	for _, name := range []string{
		"dependencies-current.jsonl", "dependencies-v2.2.1.jsonl", "dependencies-e2e.jsonl",
		"go-env.json", "buildx-version.txt", "docker-host.txt", "source-files-sha256.txt",
		"builder-cache-before-build.txt",
	} {
		writeP0P1GateTestFile(t, filepath.Join(dir, name), "evidence\n")
	}
	writeP0P1GateTestFile(t, filepath.Join(dir, "builder-cache-record-ids-before-build.txt"), "")
	var imageIndex string
	for _, image := range manifest.Images {
		imageIndex += fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s\n", image.Platform, image.Kind, image.ImageRef, image.ImageDigest, image.ImageID, image.BinarySHA256, image.Version, image.SourceCommit)
	}
	writeP0P1GateTestFile(t, filepath.Join(dir, "image-index.txt"), imageIndex)
	for _, platform := range []string{"amd64", "arm64"} {
		for _, kind := range []string{"current", "v2.2.1"} {
			prefix := kind + "-" + platform
			releaseImageID := testP0P1ReleaseCurrentImageID
			binarySHA256 := testP0P1CurrentBinarySHA256
			version := "2.3.0"
			commit := sourceCommit
			if kind == "v2.2.1" {
				releaseImageID = testP0P1ReleaseOldImageID
				binarySHA256 = testP0P1OldBinarySHA256
				version = "2.2.1"
				commit = "89abcdef0123456789abcdef0123456789abcdef"
			}
			writeP0P1GateTestJSON(t, filepath.Join(dir, prefix+"-build-metadata.json"), map[string]any{
				"containerimage.digest": releaseImageID,
			})
			writeP0P1GateTestFile(t, filepath.Join(dir, prefix+"-binary-sha256.txt"), binarySHA256+"  panacead\n")
			writeP0P1GateTestFile(t, filepath.Join(dir, prefix+"-build-args.txt"), "platform=linux/"+platform+"\nPANACEA_VERSION="+version+"\nPANACEA_COMMIT="+commit+"\n")
			writeP0P1GateTestFile(t, filepath.Join(dir, prefix+"-version.txt"), "version: "+version+"\ncommit: "+commit+"\n")
			writeP0P1GateTestJSON(t, filepath.Join(dir, prefix+"-image-inspect.json"), []map[string]any{{
				"Id": releaseImageID, "Os": "linux", "Architecture": platform,
			}})
			writeP0P1GateTestFile(t, filepath.Join(dir, prefix+"-smoke.txt"), "")
		}
		writeP0P1GateTestFile(t, filepath.Join(dir, "upgrade-"+platform+"-result.txt"), "platform=linux/"+platform+"\nresult=passed\n")
		writeP0P1GateTestFile(t, filepath.Join(dir, "upgrade-"+platform+".log"), "upgrade passed\n")
	}
	writeP0P1GateTestJSON(t, filepath.Join(dir, "functional-current-amd64-image-inspect.json"), []map[string]any{{
		"Id": testP0P1CurrentImageID, "Os": "linux", "Architecture": "amd64",
	}})
	writeP0P1GateTestJSON(t, filepath.Join(dir, "functional-v2.2.1-amd64-image-inspect.json"), []map[string]any{{
		"Id": testP0P1OldImageID, "Os": "linux", "Architecture": "amd64",
	}})
	writeP0P1GateTestFile(t, filepath.Join(dir, "functional-current-amd64-binary-sha256.txt"), testP0P1CurrentBinarySHA256+"  panacead\n")
	writeP0P1GateTestFile(t, filepath.Join(dir, "functional-v2.2.1-amd64-binary-sha256.txt"), testP0P1OldBinarySHA256+"  panacead\n")
	writeP0P1GateTestJSON(t, filepath.Join(dir, ReleaseHostImageIdentityArtifactPath), ReleaseHostImageIdentity{
		SchemaVersion: ReleaseHostImageIdentitySchemaVersion,
		HostPlatform:  "linux/amd64",
		Images:        testP0P1ReleaseHostImages(),
	})
}

func writeP0P1GateTestFile(t *testing.T, filePath, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o700))
	require.NoError(t, os.WriteFile(filePath, []byte(contents), 0o600))
}
