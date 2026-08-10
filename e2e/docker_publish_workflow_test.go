package e2e_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerPublishBuildsAndPushesIndependentlyOfE2E(t *testing.T) {
	contents, err := os.ReadFile("../.github/workflows/docker-publish.yml")
	require.NoError(t, err)
	workflow := string(contents)

	for _, contract := range []string{
		"tags: ['v*.*.*']",
		"platform: linux/amd64",
		"runner: ubuntu-24.04",
		"platform: linux/arm64",
		"runner: ubuntu-24.04-arm",
		"runs-on: ${{ matrix.runner }}",
		"actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09 # v5.1.0",
		"actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6.5.0",
		"docker/setup-buildx-action@8d2750c68a42422c14e847fe6c8ac0403b4cbd6f # v3.12.0",
		"docker/login-action@c94ce9fb468520275223c153574b00df6fe4bcc9 # v3.7.0",
		"docker/metadata-action@c299e40c65443455700f0fdfc63efafe5b349051 # v5.10.0",
		"docker/build-push-action@10e90e3645eae34f1e60eeb005ba3a3d33f178e8 # v6.19.2",
		"actions/upload-artifact@b7c566a772e6b6bfb58ed0dc250532a479d7789f # v6.0.0",
		"actions/download-artifact@37930b1c2abaa49bbe596cd826c3c89aef350131 # v7.0.0",
		"context: .",
		"platforms: ${{ matrix.platform }}",
		"push-by-digest=true",
		"labels: ${{ steps.meta.outputs.labels }}",
		"needs: build",
		"merge-multiple: true",
		"docker buildx imagetools create",
	} {
		require.Contains(t, workflow, contract)
	}
	usesCount := 0
	for _, line := range strings.Split(workflow, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "uses:") {
			continue
		}
		usesCount++
		action := strings.Fields(strings.TrimSpace(strings.TrimPrefix(line, "uses:")))[0]
		_, revision, found := strings.Cut(action, "@")
		require.True(t, found, "action must include a revision: %s", action)
		require.Regexp(t, `^[0-9a-f]{40}$`, revision, "action must use an immutable commit SHA: %s", action)
	}
	require.Equal(t, 11, usesCount)
	for _, forbidden := range []string{
		"scripts/e2e/",
		"release-hardening",
		"gate-manifest",
		".release_images",
	} {
		require.NotContains(t, strings.ToLower(workflow), forbidden)
	}
}

func TestDockerPublishPassesResolvedSourceMetadataIntoTheBuild(t *testing.T) {
	contents, err := os.ReadFile("../.github/workflows/docker-publish.yml")
	require.NoError(t, err)
	workflow := string(contents)

	for _, contract := range []string{
		"fetch-depth: 0",
		"name: Resolve source metadata",
		"id: source",
		"version=$(git describe --tags)",
		"version=${version#v}",
		"commit=$(git rev-parse HEAD)",
		`printf 'version=%s\n' "$version"`,
		`printf 'commit=%s\n' "$commit"`,
		`} >>"$GITHUB_OUTPUT"`,
		"build-args: |",
		"PANACEA_VERSION=${{ steps.source.outputs.version }}",
		"PANACEA_COMMIT=${{ steps.source.outputs.commit }}",
	} {
		require.Contains(t, workflow, contract)
	}
}
