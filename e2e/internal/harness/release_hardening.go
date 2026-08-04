package harness

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	ReleaseHardeningManifestSchemaVersion = "4"
	ReleaseHostImageIdentitySchemaVersion = "1"
	ReleaseHostImageIdentityArtifactPath  = "host-image-identity.json"
)

var (
	releaseCommitPattern      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	releaseDigestPattern      = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	releaseBareDigestPattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	releaseChecksumLine       = regexp.MustCompile(`^[0-9a-f]{64}[[:space:]]+`)
	releasePinnedFromPattern  = regexp.MustCompile(`^FROM ([^[:space:]@]+)@(sha256:[0-9a-f]{64})(?: AS ([[:alnum:]_.-]+))?$`)
	releaseExpectedPlatforms  = []string{"linux/amd64", "linux/arm64"}
	releaseExpectedImageKinds = []string{"current", "v2.2.1"}
)

// ReleaseHardeningManifest is only written after every cold-cache build,
// architecture smoke, and architecture-specific real upgrade has succeeded.
// A skipped architecture or upgrade cannot satisfy this contract.
type ReleaseHardeningManifest struct {
	SchemaVersion                 string                         `json:"schema_version"`
	RunID                         string                         `json:"run_id"`
	SourceCommit                  string                         `json:"source_commit"`
	SourceClean                   bool                           `json:"source_clean"`
	ColdGoCaches                  bool                           `json:"cold_go_caches"`
	FreshBuildKitBuilder          bool                           `json:"fresh_buildkit_builder"`
	WarmOfflineHostBuild          bool                           `json:"warm_offline_host_build"`
	WarmOfflineBuildKitBuild      bool                           `json:"warm_offline_buildkit_build"`
	DockerBuildNetwork            string                         `json:"docker_build_network"`
	Platforms                     []string                       `json:"platforms"`
	VersionAndSmoke               bool                           `json:"version_and_smoke"`
	MultiarchUpgradeCompatibility bool                           `json:"multiarch_upgrade_compatibility"`
	ImageIndex                    string                         `json:"image_index"`
	HostPlatform                  string                         `json:"host_platform"`
	HostImageIdentity             string                         `json:"host_image_identity"`
	Images                        []ReleasePlatformImageEvidence `json:"images"`
}

// ReleasePlatformImageEvidence is the publishable content contract for one
// image kind and one target architecture. The OCI manifest digest, loaded
// config ID, binary checksum, version, and source commit are all verified
// against their independently recorded artifacts before the final gate can
// consume this record.
type ReleasePlatformImageEvidence struct {
	Kind         string `json:"kind"`
	Platform     string `json:"platform"`
	ImageRef     string `json:"image_ref"`
	ImageDigest  string `json:"image_digest"`
	ImageID      string `json:"image_id"`
	BinarySHA256 string `json:"binary_sha256"`
	Version      string `json:"version"`
	SourceCommit string `json:"source_commit"`
}

func (m ReleaseHardeningManifest) Validate() error {
	if m.SchemaVersion != ReleaseHardeningManifestSchemaVersion {
		return fmt.Errorf("release hardening schema version %q, want %q", m.SchemaVersion, ReleaseHardeningManifestSchemaVersion)
	}
	if strings.TrimSpace(m.RunID) == "" {
		return errors.New("release hardening run_id is required")
	}
	if !releaseCommitPattern.MatchString(m.SourceCommit) {
		return fmt.Errorf("release hardening source_commit %q is not a full lowercase commit", m.SourceCommit)
	}
	if !m.SourceClean {
		return errors.New("release hardening source_clean must prove an unchanged HEAD worktree")
	}
	if !m.ColdGoCaches || !m.FreshBuildKitBuilder || !m.WarmOfflineHostBuild || !m.WarmOfflineBuildKitBuild {
		return errors.New("release hardening requires cold Go caches, a fresh BuildKit builder, and warm offline host and BuildKit builds")
	}
	if m.DockerBuildNetwork != "none" {
		return fmt.Errorf("release Docker build network %q, want none", m.DockerBuildNetwork)
	}
	platforms := append([]string(nil), m.Platforms...)
	sort.Strings(platforms)
	wantPlatforms := append([]string(nil), releaseExpectedPlatforms...)
	sort.Strings(wantPlatforms)
	if strings.Join(platforms, ",") != strings.Join(wantPlatforms, ",") {
		return fmt.Errorf("release hardening platforms %v, want exactly %v", m.Platforms, releaseExpectedPlatforms)
	}
	if !m.VersionAndSmoke {
		return errors.New("release hardening version_and_smoke must pass")
	}
	if !m.MultiarchUpgradeCompatibility {
		return errors.New("release hardening multiarch upgrade compatibility was not executed successfully")
	}
	if m.ImageIndex != "image-index.txt" {
		return fmt.Errorf("release hardening image index %q, want image-index.txt", m.ImageIndex)
	}
	if !containsReleasePlatform(m.HostPlatform) {
		return fmt.Errorf("release hardening host platform %q is not one of %v", m.HostPlatform, releaseExpectedPlatforms)
	}
	if m.HostImageIdentity != ReleaseHostImageIdentityArtifactPath {
		return fmt.Errorf("release hardening host image identity %q, want %q", m.HostImageIdentity, ReleaseHostImageIdentityArtifactPath)
	}
	if err := validateReleasePlatformImages(m.Images, m.RunID, m.SourceCommit); err != nil {
		return err
	}
	return nil
}

func validateReleasePlatformImages(images []ReleasePlatformImageEvidence, runID, currentCommit string) error {
	if len(images) != len(releaseExpectedPlatforms)*len(releaseExpectedImageKinds) {
		return fmt.Errorf("release hardening has %d platform images, want %d", len(images), len(releaseExpectedPlatforms)*len(releaseExpectedImageKinds))
	}
	seen := make(map[string]ReleasePlatformImageEvidence, len(images))
	oldCommit := ""
	for _, image := range images {
		if !containsReleasePlatform(image.Platform) || !containsReleaseImageKind(image.Kind) {
			return fmt.Errorf("release hardening image has invalid kind/platform %q/%q", image.Kind, image.Platform)
		}
		key := image.Kind + "|" + image.Platform
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("release hardening contains duplicate platform image %s", key)
		}
		suffix := strings.TrimPrefix(image.Platform, "linux/")
		wantRef := fmt.Sprintf("panacea-e2e-release-%s-%s:%s", image.Kind, suffix, runID)
		if image.ImageRef != wantRef {
			return fmt.Errorf("release hardening image %s ref %q, want %q", key, image.ImageRef, wantRef)
		}
		if !releaseDigestPattern.MatchString(image.ImageDigest) || !releaseDigestPattern.MatchString(image.ImageID) {
			return fmt.Errorf("release hardening image %s requires OCI and config sha256 digests", key)
		}
		if !releaseBareDigestPattern.MatchString(image.BinarySHA256) {
			return fmt.Errorf("release hardening image %s binary checksum %q is invalid", key, image.BinarySHA256)
		}
		if !releaseCommitPattern.MatchString(image.SourceCommit) || strings.TrimSpace(image.Version) == "" {
			return fmt.Errorf("release hardening image %s requires a version and full source commit", key)
		}
		if image.Kind == "current" {
			if image.SourceCommit != currentCommit {
				return fmt.Errorf("release hardening current image %s commit %s, want %s", image.Platform, image.SourceCommit, currentCommit)
			}
		} else if oldCommit == "" {
			oldCommit = image.SourceCommit
		} else if image.SourceCommit != oldCommit {
			return fmt.Errorf("release hardening compatibility image commits differ: %s and %s", oldCommit, image.SourceCommit)
		}
		seen[key] = image
	}
	for _, platform := range releaseExpectedPlatforms {
		for _, kind := range releaseExpectedImageKinds {
			if _, ok := seen[kind+"|"+platform]; !ok {
				return fmt.Errorf("release hardening is missing platform image %s/%s", kind, platform)
			}
		}
	}
	return nil
}

// ReleaseHostImageIdentity ties each content-addressed image used by the
// functional suites to the independently rebuilt release image for the Docker
// daemon's native platform. Image config IDs may differ between builds, so the
// panacead binary checksum is the fail-closed equivalence contract.
type ReleaseHostImageIdentity struct {
	SchemaVersion string                          `json:"schema_version"`
	HostPlatform  string                          `json:"host_platform"`
	Images        []ReleaseHostImageIdentityEntry `json:"images"`
}

type ReleaseHostImageIdentityEntry struct {
	Kind                   string `json:"kind"`
	FunctionalImageRef     string `json:"functional_image_ref"`
	FunctionalImageID      string `json:"functional_image_id"`
	FunctionalBinarySHA256 string `json:"functional_binary_sha256"`
	ReleaseImageRef        string `json:"release_image_ref"`
	ReleaseImageID         string `json:"release_image_id"`
	ReleaseBinarySHA256    string `json:"release_binary_sha256"`
}

func (i ReleaseHostImageIdentity) Validate() error {
	if i.SchemaVersion != ReleaseHostImageIdentitySchemaVersion {
		return fmt.Errorf("release host image identity schema version %q, want %q", i.SchemaVersion, ReleaseHostImageIdentitySchemaVersion)
	}
	if !containsReleasePlatform(i.HostPlatform) {
		return fmt.Errorf("release host image identity platform %q is not one of %v", i.HostPlatform, releaseExpectedPlatforms)
	}
	if len(i.Images) != len(releaseExpectedImageKinds) {
		return fmt.Errorf("release host image identity has %d images, want %d", len(i.Images), len(releaseExpectedImageKinds))
	}
	seen := make(map[string]ReleaseHostImageIdentityEntry, len(i.Images))
	for _, image := range i.Images {
		if _, duplicate := seen[image.Kind]; duplicate {
			return fmt.Errorf("release host image identity contains duplicate kind %q", image.Kind)
		}
		if !containsReleaseImageKind(image.Kind) {
			return fmt.Errorf("release host image identity contains unexpected kind %q", image.Kind)
		}
		if strings.TrimSpace(image.FunctionalImageRef) == "" || strings.TrimSpace(image.ReleaseImageRef) == "" {
			return fmt.Errorf("release host image identity %q requires functional and release image references", image.Kind)
		}
		if !releaseDigestPattern.MatchString(image.FunctionalImageID) || !releaseDigestPattern.MatchString(image.ReleaseImageID) {
			return fmt.Errorf("release host image identity %q requires sha256 functional and release image IDs", image.Kind)
		}
		if !releaseBareDigestPattern.MatchString(image.FunctionalBinarySHA256) || !releaseBareDigestPattern.MatchString(image.ReleaseBinarySHA256) {
			return fmt.Errorf("release host image identity %q requires lowercase panacead SHA-256 checksums", image.Kind)
		}
		if image.FunctionalBinarySHA256 != image.ReleaseBinarySHA256 {
			return fmt.Errorf("release host image identity %q functional binary %s differs from release binary %s", image.Kind, image.FunctionalBinarySHA256, image.ReleaseBinarySHA256)
		}
		seen[image.Kind] = image
	}
	for _, kind := range releaseExpectedImageKinds {
		if _, ok := seen[kind]; !ok {
			return fmt.Errorf("release host image identity is missing kind %q", kind)
		}
	}
	if seen["current"].FunctionalImageID == seen["v2.2.1"].FunctionalImageID {
		return errors.New("release host image identity current and v2.2.1 functional image IDs must differ")
	}
	if seen["current"].ReleaseImageID == seen["v2.2.1"].ReleaseImageID {
		return errors.New("release host image identity current and v2.2.1 release image IDs must differ")
	}
	return nil
}

func (i ReleaseHostImageIdentity) image(kind string) (ReleaseHostImageIdentityEntry, bool) {
	for _, image := range i.Images {
		if image.Kind == kind {
			return image, true
		}
	}
	return ReleaseHostImageIdentityEntry{}, false
}

func containsReleasePlatform(platform string) bool {
	for _, expected := range releaseExpectedPlatforms {
		if platform == expected {
			return true
		}
	}
	return false
}

func containsReleaseImageKind(kind string) bool {
	for _, expected := range releaseExpectedImageKinds {
		if kind == expected {
			return true
		}
	}
	return false
}

type ReleasePinnedBaseImage struct {
	Reference string `json:"reference"`
	Digest    string `json:"digest"`
	Stage     string `json:"stage,omitempty"`
}

// ParseReleasePinnedBaseImages rejects floating tags and digest-less stages.
func ParseReleasePinnedBaseImages(contents []byte) ([]ReleasePinnedBaseImage, error) {
	var images []ReleasePinnedBaseImage
	scanner := bufio.NewScanner(strings.NewReader(string(contents)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "FROM ") {
			continue
		}
		match := releasePinnedFromPattern.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("Dockerfile base is not pinned by a sha256 digest: %q", line)
		}
		images = append(images, ReleasePinnedBaseImage{Reference: match[1], Digest: match[2], Stage: match[3]})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(images) < 2 {
		return nil, fmt.Errorf("Dockerfile has %d pinned base images, want at least two", len(images))
	}
	if images[0].Stage != "build-env" || !strings.HasPrefix(images[0].Reference, "golang:1.23.12-") {
		return nil, fmt.Errorf("first release base must be the staged Go 1.23.12 builder, got %+v", images[0])
	}
	return images, nil
}

// ValidateReleaseHardeningArtifact verifies the durable evidence index and
// every architecture-specific result file. It intentionally does not infer a
// pass from build logs: explicit result files are emitted only after commands
// exit successfully.
func ValidateReleaseHardeningArtifact(manifestPath string) error {
	manifestPath = filepath.Clean(manifestPath)
	if !filepath.IsAbs(manifestPath) {
		return fmt.Errorf("release hardening manifest path must be absolute: %s", manifestPath)
	}
	contents, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var manifest ReleaseHardeningManifest
	if err := json.Unmarshal(contents, &manifest); err != nil {
		return fmt.Errorf("decode release hardening manifest: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return err
	}
	dir := filepath.Dir(manifestPath)
	for _, name := range []string{"source-status.txt", "source-diff.patch", "source-status-final.txt", "source-diff-final.patch"} {
		if err := requireEmptyReleaseArtifact(dir, name); err != nil {
			return fmt.Errorf("clean release source evidence: %w", err)
		}
	}
	for _, name := range []string{"source-commit.txt", "source-commit-final.txt"} {
		contents, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(contents)) != manifest.SourceCommit {
			return fmt.Errorf("release source commit artifact %s is %q, want %s", name, strings.TrimSpace(string(contents)), manifest.SourceCommit)
		}
	}

	baseImages, err := os.ReadFile(filepath.Join(dir, "base-images.txt"))
	if err != nil {
		return err
	}
	if _, err := ParseReleasePinnedBaseImages(baseImages); err != nil {
		return err
	}
	for _, dependencyFile := range []string{
		"dependencies-current.jsonl",
		"dependencies-v2.2.1.jsonl",
		"dependencies-e2e.jsonl",
		"go-env.json",
		"buildx-version.txt",
		"docker-host.txt",
		"source-files-sha256.txt",
		"builder-cache-before-build.txt",
		"builder-networks-before-offline.txt",
		"warm-offline-buildkit-contract.txt",
	} {
		if err := requireNonemptyReleaseArtifact(dir, dependencyFile); err != nil {
			return err
		}
	}
	for _, emptyEvidence := range []string{
		"builder-cache-record-ids-before-build.txt",
		"builder-networks-after-offline.txt",
	} {
		info, err := os.Stat(filepath.Join(dir, emptyEvidence))
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() != 0 {
			return fmt.Errorf("release artifact %s must be an empty regular file", emptyEvidence)
		}
	}

	indexContents, err := os.ReadFile(filepath.Join(dir, manifest.ImageIndex))
	if err != nil {
		return err
	}
	manifestImages := make(map[string]ReleasePlatformImageEvidence, len(manifest.Images))
	for _, image := range manifest.Images {
		manifestImages[image.Kind+"|"+image.Platform] = image
	}
	indexedImages := make(map[string]struct{}, len(manifest.Images))
	for _, line := range strings.Split(strings.TrimSpace(string(indexContents)), "\n") {
		parts := strings.Split(line, "|")
		if len(parts) != 8 {
			return fmt.Errorf("invalid release image index line %q", line)
		}
		indexed := ReleasePlatformImageEvidence{
			Platform: parts[0], Kind: parts[1], ImageRef: parts[2], ImageDigest: parts[3],
			ImageID: parts[4], BinarySHA256: parts[5], Version: parts[6], SourceCommit: parts[7],
		}
		key := indexed.Kind + "|" + indexed.Platform
		if _, duplicate := indexedImages[key]; duplicate {
			return fmt.Errorf("duplicate release image index entry %s", key)
		}
		if want, ok := manifestImages[key]; !ok || indexed != want {
			return fmt.Errorf("release image index entry %s does not match manifest evidence", key)
		}
		indexedImages[key] = struct{}{}
	}

	for _, platform := range releaseExpectedPlatforms {
		suffix := strings.TrimPrefix(platform, "linux/")
		for _, kind := range releaseExpectedImageKinds {
			record := manifestImages[kind+"|"+platform]
			prefix := kind + "-" + suffix
			metadataContents, err := os.ReadFile(filepath.Join(dir, prefix+"-build-metadata.json"))
			if err != nil {
				return err
			}
			var metadata map[string]any
			if err := json.Unmarshal(metadataContents, &metadata); err != nil {
				return fmt.Errorf("decode %s build metadata: %w", prefix, err)
			}
			digest, _ := metadata["containerimage.digest"].(string)
			if digest != record.ImageDigest {
				return fmt.Errorf("%s OCI image digest %q, want manifest %q", prefix, digest, record.ImageDigest)
			}
			checksum, err := readReleaseChecksum(filepath.Join(dir, prefix+"-binary-sha256.txt"))
			if err != nil {
				return err
			}
			if checksum != record.BinarySHA256 {
				return fmt.Errorf("%s binary checksum %s, want manifest %s", prefix, checksum, record.BinarySHA256)
			}
			for _, suffixName := range []string{"build-args.txt", "version.txt", "image-inspect.json"} {
				if err := requireNonemptyReleaseArtifact(dir, prefix+"-"+suffixName); err != nil {
					return err
				}
			}
			inspect, err := readReleaseDockerImageInspect(filepath.Join(dir, prefix+"-image-inspect.json"))
			if err != nil {
				return err
			}
			if err := requireReleaseInspectIdentity(inspect, platform, record.ImageID); err != nil {
				return fmt.Errorf("validate %s image inspect: %w", prefix, err)
			}
			buildArgs, err := os.ReadFile(filepath.Join(dir, prefix+"-build-args.txt"))
			if err != nil {
				return err
			}
			for _, contract := range []string{"platform=" + platform + "\n", "PANACEA_VERSION=" + record.Version + "\n", "PANACEA_COMMIT=" + record.SourceCommit + "\n"} {
				if !strings.Contains(string(buildArgs), contract) {
					return fmt.Errorf("%s build args are missing %q", prefix, strings.TrimSpace(contract))
				}
			}
			version, err := os.ReadFile(filepath.Join(dir, prefix+"-version.txt"))
			if err != nil {
				return err
			}
			if !strings.Contains(string(version), "version: "+record.Version+"\n") || !strings.Contains(string(version), "commit: "+record.SourceCommit+"\n") {
				return fmt.Errorf("%s version evidence does not match manifest version/commit", prefix)
			}
			if _, err := os.Stat(filepath.Join(dir, prefix+"-smoke.txt")); err != nil {
				return err
			}
			warmPrefix := "warm-offline-" + prefix
			warmMetadataContents, err := os.ReadFile(filepath.Join(dir, warmPrefix+"-build-metadata.json"))
			if err != nil {
				return err
			}
			var warmMetadata map[string]any
			if err := json.Unmarshal(warmMetadataContents, &warmMetadata); err != nil {
				return fmt.Errorf("decode %s build metadata: %w", warmPrefix, err)
			}
			warmDigest, _ := warmMetadata["containerimage.digest"].(string)
			if !releaseDigestPattern.MatchString(warmDigest) {
				return fmt.Errorf("%s OCI image digest %q is invalid", warmPrefix, warmDigest)
			}
			if warmDigest != digest {
				return fmt.Errorf("%s OCI image digest %s differs from cold build %s", warmPrefix, warmDigest, digest)
			}
			for _, suffixName := range []string{"build.log", "image-inspect.json", "version.txt"} {
				if err := requireNonemptyReleaseArtifact(dir, warmPrefix+"-"+suffixName); err != nil {
					return err
				}
			}
			warmInspect, err := readReleaseDockerImageInspect(filepath.Join(dir, warmPrefix+"-image-inspect.json"))
			if err != nil {
				return err
			}
			if err := requireReleaseInspectIdentity(warmInspect, platform, record.ImageID); err != nil {
				return fmt.Errorf("validate %s image inspect: %w", warmPrefix, err)
			}
			warmVersion, err := os.ReadFile(filepath.Join(dir, warmPrefix+"-version.txt"))
			if err != nil {
				return err
			}
			if !strings.Contains(string(warmVersion), "version: "+record.Version+"\n") || !strings.Contains(string(warmVersion), "commit: "+record.SourceCommit+"\n") {
				return fmt.Errorf("%s version evidence does not match manifest version/commit", warmPrefix)
			}
		}

		resultContents, err := os.ReadFile(filepath.Join(dir, "upgrade-"+suffix+"-result.txt"))
		if err != nil {
			return err
		}
		result := string(resultContents)
		if !strings.Contains(result, "platform="+platform+"\n") || !strings.Contains(result, "result=passed\n") {
			return fmt.Errorf("%s upgrade result is not an explicit pass", platform)
		}
		if err := requireNonemptyReleaseArtifact(dir, "upgrade-"+suffix+".log"); err != nil {
			return err
		}
	}
	if len(indexedImages) != len(manifest.Images) {
		return fmt.Errorf("release image index contains %d images, want %d", len(indexedImages), len(manifest.Images))
	}
	if _, err := readReleaseHostImageIdentity(dir, manifest); err != nil {
		return err
	}
	return nil
}

type releaseDockerImageInspect struct {
	ID           string `json:"Id"`
	OS           string `json:"Os"`
	Architecture string `json:"Architecture"`
}

func readReleaseHostImageIdentity(dir string, manifest ReleaseHardeningManifest) (ReleaseHostImageIdentity, error) {
	var zero ReleaseHostImageIdentity
	contents, err := os.ReadFile(filepath.Join(dir, manifest.HostImageIdentity))
	if err != nil {
		return zero, err
	}
	var identity ReleaseHostImageIdentity
	if err := json.Unmarshal(contents, &identity); err != nil {
		return zero, fmt.Errorf("decode release host image identity: %w", err)
	}
	if err := identity.Validate(); err != nil {
		return zero, err
	}
	if identity.HostPlatform != manifest.HostPlatform {
		return zero, fmt.Errorf("release host image identity platform %s, want manifest platform %s", identity.HostPlatform, manifest.HostPlatform)
	}
	suffix := strings.TrimPrefix(identity.HostPlatform, "linux/")
	for _, image := range identity.Images {
		functionalPrefix := "functional-" + image.Kind + "-" + suffix
		releasePrefix := image.Kind + "-" + suffix
		functionalInspect, err := readReleaseDockerImageInspect(filepath.Join(dir, functionalPrefix+"-image-inspect.json"))
		if err != nil {
			return zero, fmt.Errorf("validate %s image inspect: %w", functionalPrefix, err)
		}
		if err := requireReleaseInspectIdentity(functionalInspect, identity.HostPlatform, image.FunctionalImageID); err != nil {
			return zero, fmt.Errorf("validate %s image inspect: %w", functionalPrefix, err)
		}
		releaseInspect, err := readReleaseDockerImageInspect(filepath.Join(dir, releasePrefix+"-image-inspect.json"))
		if err != nil {
			return zero, fmt.Errorf("validate %s image inspect: %w", releasePrefix, err)
		}
		if err := requireReleaseInspectIdentity(releaseInspect, identity.HostPlatform, image.ReleaseImageID); err != nil {
			return zero, fmt.Errorf("validate %s image inspect: %w", releasePrefix, err)
		}
		functionalChecksum, err := readReleaseChecksum(filepath.Join(dir, functionalPrefix+"-binary-sha256.txt"))
		if err != nil {
			return zero, err
		}
		if functionalChecksum != image.FunctionalBinarySHA256 {
			return zero, fmt.Errorf("%s binary checksum %s, want identity checksum %s", functionalPrefix, functionalChecksum, image.FunctionalBinarySHA256)
		}
		releaseChecksum, err := readReleaseChecksum(filepath.Join(dir, releasePrefix+"-binary-sha256.txt"))
		if err != nil {
			return zero, err
		}
		if releaseChecksum != image.ReleaseBinarySHA256 {
			return zero, fmt.Errorf("%s binary checksum %s, want identity checksum %s", releasePrefix, releaseChecksum, image.ReleaseBinarySHA256)
		}
	}
	return identity, nil
}

func readReleaseDockerImageInspect(path string) (releaseDockerImageInspect, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return releaseDockerImageInspect{}, err
	}
	var inspections []releaseDockerImageInspect
	if err := json.Unmarshal(contents, &inspections); err != nil {
		return releaseDockerImageInspect{}, fmt.Errorf("decode Docker image inspect: %w", err)
	}
	if len(inspections) != 1 {
		return releaseDockerImageInspect{}, fmt.Errorf("Docker image inspect contains %d images, want one", len(inspections))
	}
	return inspections[0], nil
}

func requireReleaseInspectIdentity(inspect releaseDockerImageInspect, platform, imageID string) error {
	if inspect.ID != imageID {
		return fmt.Errorf("Docker image ID %q, want %q", inspect.ID, imageID)
	}
	if inspect.OS+"/"+inspect.Architecture != platform {
		return fmt.Errorf("Docker image platform %s/%s, want %s", inspect.OS, inspect.Architecture, platform)
	}
	return nil
}

func readReleaseChecksum(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(contents))
	if len(fields) < 2 || !releaseBareDigestPattern.MatchString(fields[0]) {
		return "", fmt.Errorf("release binary checksum %s is invalid", filepath.Base(path))
	}
	return fields[0], nil
}

func requireNonemptyReleaseArtifact(dir, name string) error {
	info, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("release artifact %s is empty or not a regular file", name)
	}
	return nil
}

func requireEmptyReleaseArtifact(dir, name string) error {
	info, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() != 0 {
		return fmt.Errorf("release artifact %s must be an empty regular file", name)
	}
	return nil
}
