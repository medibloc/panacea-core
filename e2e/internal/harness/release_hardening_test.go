package harness

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReleaseHardeningManifestRequiresBothArchitecturesAndRealUpgrades(t *testing.T) {
	valid := ReleaseHardeningManifest{
		SchemaVersion:                 ReleaseHardeningManifestSchemaVersion,
		RunID:                         "release-20260804000000-1",
		SourceCommit:                  "0123456789abcdef0123456789abcdef01234567",
		SourceClean:                   true,
		ColdGoCaches:                  true,
		FreshBuildKitBuilder:          true,
		WarmOfflineHostBuild:          true,
		WarmOfflineBuildKitBuild:      true,
		DockerBuildNetwork:            "none",
		Platforms:                     []string{"linux/arm64", "linux/amd64"},
		VersionAndSmoke:               true,
		MultiarchUpgradeCompatibility: true,
		ImageIndex:                    "image-index.txt",
		HostPlatform:                  "linux/amd64",
		HostImageIdentity:             ReleaseHostImageIdentityArtifactPath,
		Images:                        testP0P1ReleasePlatformImages("0123456789abcdef0123456789abcdef01234567"),
	}
	require.NoError(t, valid.Validate())

	missingArchitecture := valid
	missingArchitecture.Platforms = []string{"linux/arm64"}
	require.ErrorContains(t, missingArchitecture.Validate(), "exactly")

	notRun := valid
	notRun.MultiarchUpgradeCompatibility = false
	require.ErrorContains(t, notRun.Validate(), "was not executed successfully")

	networkedBuild := valid
	networkedBuild.DockerBuildNetwork = "default"
	require.ErrorContains(t, networkedBuild.Validate(), "want none")

	registryDependentBuild := valid
	registryDependentBuild.WarmOfflineBuildKitBuild = false
	require.ErrorContains(t, registryDependentBuild.Validate(), "BuildKit")

	dirtySource := valid
	dirtySource.SourceClean = false
	require.ErrorContains(t, dirtySource.Validate(), "unchanged HEAD")

	missingHostIdentity := valid
	missingHostIdentity.HostImageIdentity = ""
	require.ErrorContains(t, missingHostIdentity.Validate(), "host image identity")

	wrongNonHostCommit := valid
	wrongNonHostCommit.Images = append([]ReleasePlatformImageEvidence(nil), valid.Images...)
	for index := range wrongNonHostCommit.Images {
		if wrongNonHostCommit.Images[index].Kind == "current" && wrongNonHostCommit.Images[index].Platform == "linux/arm64" {
			wrongNonHostCommit.Images[index].SourceCommit = "89abcdef0123456789abcdef0123456789abcdef"
		}
	}
	require.ErrorContains(t, wrongNonHostCommit.Validate(), "want")
}

func TestReleaseHostImageIdentityRequiresMatchingFunctionalAndReleaseBinaries(t *testing.T) {
	valid := ReleaseHostImageIdentity{
		SchemaVersion: ReleaseHostImageIdentitySchemaVersion,
		HostPlatform:  "linux/amd64",
		Images: []ReleaseHostImageIdentityEntry{
			{
				Kind: "current", FunctionalImageRef: "panacea-e2e-current:local",
				FunctionalImageID:      "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				FunctionalBinarySHA256: "1111111111111111111111111111111111111111111111111111111111111111",
				ReleaseImageRef:        "panacea-e2e-release-current-amd64:run",
				ReleaseImageID:         "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				ReleaseBinarySHA256:    "1111111111111111111111111111111111111111111111111111111111111111",
			},
			{
				Kind: "v2.2.1", FunctionalImageRef: "panacea-e2e-v2.2.1:local",
				FunctionalImageID:      "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				FunctionalBinarySHA256: "2222222222222222222222222222222222222222222222222222222222222222",
				ReleaseImageRef:        "panacea-e2e-release-v2.2.1-amd64:run",
				ReleaseImageID:         "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				ReleaseBinarySHA256:    "2222222222222222222222222222222222222222222222222222222222222222",
			},
		},
	}
	require.NoError(t, valid.Validate())

	mismatch := valid
	mismatch.Images = append([]ReleaseHostImageIdentityEntry(nil), valid.Images...)
	mismatch.Images[0].ReleaseBinarySHA256 = "3333333333333333333333333333333333333333333333333333333333333333"
	require.ErrorContains(t, mismatch.Validate(), "differs")

	duplicate := valid
	duplicate.Images = append([]ReleaseHostImageIdentityEntry(nil), valid.Images...)
	duplicate.Images[1].Kind = "current"
	require.ErrorContains(t, duplicate.Validate(), "duplicate")
}

func TestParseReleasePinnedBaseImagesRejectsFloatingOrWrongGoBuilder(t *testing.T) {
	valid := []byte(`FROM golang:1.23.12-bullseye@sha256:161b8513c09cbfa4c174fd32e46eddc5eddf487a43958b9cf8b07d628e9e0f85 AS build-env
FROM debian:bullseye-slim@sha256:cba95a21c96c1f5fc2470081829363eed57706634f7dc26e8c6712934303d57a
`)
	images, err := ParseReleasePinnedBaseImages(valid)
	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Equal(t, "build-env", images[0].Stage)

	_, err = ParseReleasePinnedBaseImages([]byte("FROM golang:1.23.12-bullseye AS build-env\nFROM debian:bullseye-slim\n"))
	require.ErrorContains(t, err, "not pinned")

	wrongGo := []byte(`FROM golang:1.24.0-bullseye@sha256:161b8513c09cbfa4c174fd32e46eddc5eddf487a43958b9cf8b07d628e9e0f85 AS build-env
FROM debian:bullseye-slim@sha256:cba95a21c96c1f5fc2470081829363eed57706634f7dc26e8c6712934303d57a
`)
	_, err = ParseReleasePinnedBaseImages(wrongGo)
	require.ErrorContains(t, err, "Go 1.23.12")
}

// TestValidateReleaseHardeningArtifact is normally skipped. The release
// runner sets the path only after all architecture builds and upgrades have
// exited successfully, then invokes this same contract with networking off.
func TestValidateReleaseHardeningArtifact(t *testing.T) {
	manifestPath := os.Getenv("PANACEA_E2E_RELEASE_MANIFEST")
	if manifestPath == "" {
		t.Skip("PANACEA_E2E_RELEASE_MANIFEST is set by scripts/e2e/release-hardening.sh")
	}
	require.NoError(t, ValidateReleaseHardeningArtifact(manifestPath))
}
