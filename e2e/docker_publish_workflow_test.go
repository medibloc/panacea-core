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
		"branches: [main]",
		"tags: ['v*.*.*']",
		"actions/checkout@v4",
		"docker/setup-buildx-action@v3",
		"docker/login-action@v3",
		"docker/metadata-action@v5",
		"docker/build-push-action@v6",
		"context: .",
		"push: true",
		"tags: ${{ steps.meta.outputs.tags }}",
		"labels: ${{ steps.meta.outputs.labels }}",
	} {
		require.Contains(t, workflow, contract)
	}
	for _, forbidden := range []string{
		"scripts/e2e/",
		"release-hardening",
		"gate-manifest",
		".release_images",
		"imagetools create",
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
		`printf 'version=%s\n' "$version" >>"$GITHUB_OUTPUT"`,
		`printf 'commit=%s\n' "$commit" >>"$GITHUB_OUTPUT"`,
		"build-args: |",
		"PANACEA_VERSION=${{ steps.source.outputs.version }}",
		"PANACEA_COMMIT=${{ steps.source.outputs.commit }}",
	} {
		require.Contains(t, workflow, contract)
	}
}
