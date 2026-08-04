package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
	"github.com/stretchr/testify/require"
)

var upgradeV221ExpectedModuleVersions = map[string]uint64{
	"aol":          1,
	"auth":         4,
	"authz":        2,
	"bank":         4,
	"burn":         1,
	"capability":   1,
	"consensus":    1,
	"crisis":       2,
	"did":          1,
	"distribution": 3,
	"evidence":     1,
	"feegrant":     2,
	"genutil":      1,
	"gov":          4,
	"group":        2,
	"ibc":          4,
	"mint":         2,
	"params":       1,
	"pnft":         1,
	"slashing":     3,
	"staking":      4,
	"transfer":     3,
	"upgrade":      2,
	"vesting":      1,
}

var upgradeCurrentExpectedModuleVersions = map[string]uint64{
	"aol":          1,
	"auth":         5,
	"authz":        2,
	"bank":         4,
	"burn":         1,
	"capability":   1,
	"consensus":    1,
	"crisis":       2,
	"did":          1,
	"distribution": 3,
	"evidence":     1,
	"feegrant":     2,
	"genutil":      1,
	"gov":          5,
	"group":        2,
	"ibc":          6,
	"mint":         2,
	"nft":          1,
	"params":       1,
	"pnft":         1,
	"slashing":     4,
	"staking":      5,
	"transfer":     5,
	"upgrade":      2,
	"vesting":      1,
}

type upgradeModuleVersionMismatch struct {
	Name     string `json:"name"`
	Expected uint64 `json:"expected"`
	Actual   uint64 `json:"actual"`
}

type upgradeModuleVersionComparison struct {
	Missing    []string                       `json:"missing"`
	Unexpected []string                       `json:"unexpected"`
	Mismatched []upgradeModuleVersionMismatch `json:"mismatched"`
}

type upgradeModuleVersionEvidence struct {
	Phase       string                         `json:"phase"`
	RecordedAt  time.Time                      `json:"recorded_at"`
	Expected    map[string]uint64              `json:"expected"`
	Actual      map[string]uint64              `json:"actual"`
	Comparison  upgradeModuleVersionComparison `json:"comparison"`
	DecodeError string                         `json:"decode_error,omitempty"`
	RawResponse any                            `json:"raw_response"`
}

func compareExactUpgradeModuleVersions(expected, actual map[string]uint64) upgradeModuleVersionComparison {
	comparison := upgradeModuleVersionComparison{}
	for name, expectedVersion := range expected {
		actualVersion, present := actual[name]
		if !present {
			comparison.Missing = append(comparison.Missing, name)
			continue
		}
		if actualVersion != expectedVersion {
			comparison.Mismatched = append(comparison.Mismatched, upgradeModuleVersionMismatch{
				Name:     name,
				Expected: expectedVersion,
				Actual:   actualVersion,
			})
		}
	}
	for name := range actual {
		if _, present := expected[name]; !present {
			comparison.Unexpected = append(comparison.Unexpected, name)
		}
	}
	sort.Strings(comparison.Missing)
	sort.Strings(comparison.Unexpected)
	sort.Slice(comparison.Mismatched, func(i, j int) bool {
		return comparison.Mismatched[i].Name < comparison.Mismatched[j].Name
	})
	return comparison
}

func (comparison upgradeModuleVersionComparison) err(phase string) error {
	if len(comparison.Missing) == 0 && len(comparison.Unexpected) == 0 && len(comparison.Mismatched) == 0 {
		return nil
	}
	return fmt.Errorf(
		"%s module-version map differs: missing=%v unexpected=%v mismatched=%v",
		phase,
		comparison.Missing,
		comparison.Unexpected,
		comparison.Mismatched,
	)
}

func captureExactUpgradeModuleVersions(
	ctx context.Context,
	network *harness.Network,
	phase string,
	artifactPath string,
	binaryVersion string,
	expected map[string]uint64,
) (map[string]uint64, error) {
	if strings.TrimSpace(phase) == "" || strings.TrimSpace(artifactPath) == "" {
		return nil, fmt.Errorf("module-version phase and artifact path are required")
	}
	if len(expected) == 0 {
		return nil, fmt.Errorf("%s expected module-version map is empty", phase)
	}
	command, err := upgradeModuleVersionsQueryCommand(binaryVersion)
	if err != nil {
		return nil, fmt.Errorf("%s module-version query contract: %w", phase, err)
	}
	raw, err := network.FullNodeCLIQuery(ctx, "upgrade-"+phase+"-module-versions", command...)
	if err != nil {
		return nil, err
	}
	actual, err := decodeUpgradeModuleVersions(raw)
	if err != nil {
		evidence := upgradeModuleVersionEvidence{
			Phase:       phase,
			RecordedAt:  time.Now().UTC(),
			Expected:    expected,
			DecodeError: err.Error(),
			RawResponse: jsonRawOrString(raw),
		}
		if artifactErr := network.WriteArtifactJSON(artifactPath, evidence); artifactErr != nil {
			return nil, fmt.Errorf("decode module versions: %v; record evidence: %w", err, artifactErr)
		}
		return nil, err
	}
	comparison := compareExactUpgradeModuleVersions(expected, actual)
	var rawResponse any = string(raw)
	if len(raw) > 0 {
		rawResponse = jsonRawOrString(raw)
	}
	evidence := upgradeModuleVersionEvidence{
		Phase:       phase,
		RecordedAt:  time.Now().UTC(),
		Expected:    expected,
		Actual:      actual,
		Comparison:  comparison,
		RawResponse: rawResponse,
	}
	if err := network.WriteArtifactJSON(artifactPath, evidence); err != nil {
		return nil, err
	}
	if err := comparison.err(phase); err != nil {
		return actual, err
	}
	return actual, nil
}

func upgradeModuleVersionsQueryCommand(binaryVersion string) ([]string, error) {
	switch binaryVersion {
	case "2.2.1":
		// Cosmos SDK v0.47 exposed the command with an underscore.
		return []string{"upgrade", "module_versions"}, nil
	case upgradeBinaryVersion:
		// Cosmos SDK v0.50 renamed it to the kebab-case command.
		return []string{"upgrade", "module-versions"}, nil
	default:
		return nil, fmt.Errorf("unsupported panacead version %q", binaryVersion)
	}
}

func jsonRawOrString(raw []byte) any {
	if json.Valid(raw) {
		return append(json.RawMessage(nil), raw...)
	}
	return string(raw)
}

func TestCompareExactUpgradeModuleVersionsRejectsEveryMapDrift(t *testing.T) {
	t.Parallel()
	comparison := compareExactUpgradeModuleVersions(
		map[string]uint64{"auth": 5, "bank": 4, "nft": 1},
		map[string]uint64{"auth": 4, "bank": 4, "unexpected": 9},
	)
	require.Equal(t, []string{"nft"}, comparison.Missing)
	require.Equal(t, []string{"unexpected"}, comparison.Unexpected)
	require.Equal(t, []upgradeModuleVersionMismatch{{Name: "auth", Expected: 5, Actual: 4}}, comparison.Mismatched)
	require.ErrorContains(t, comparison.err("post-upgrade"), "missing=[nft]")
	require.NoError(t, compareExactUpgradeModuleVersions(
		upgradeCurrentExpectedModuleVersions,
		upgradeCurrentExpectedModuleVersions,
	).err("exact"))
}

func TestUpgradeModuleVersionsQueryCommandIsVersionSpecific(t *testing.T) {
	t.Parallel()

	oldCommand, err := upgradeModuleVersionsQueryCommand("2.2.1")
	require.NoError(t, err)
	require.Equal(t, []string{"upgrade", "module_versions"}, oldCommand)

	currentCommand, err := upgradeModuleVersionsQueryCommand(upgradeBinaryVersion)
	require.NoError(t, err)
	require.Equal(t, []string{"upgrade", "module-versions"}, currentCommand)

	_, err = upgradeModuleVersionsQueryCommand("2.2.2")
	require.ErrorContains(t, err, "unsupported panacead version")
}
