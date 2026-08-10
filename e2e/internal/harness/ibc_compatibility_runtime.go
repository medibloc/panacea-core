package harness

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const (
	ibcBinaryChecksumMaxBytes           = 512 << 20
	hermesUpgradeInvarianceArtifactPath = "ibc/hermes/upgrade-invariance.json"
	hermesPostChannelConfigArtifactPath = "ibc/hermes/config-post-channel.toml"
	hermesPostUpgradeConfigArtifactPath = "ibc/hermes/config-post-upgrade.toml"
	hermesExpectedCompatibilityMode     = "0.37"
)

type hermesUpgradeInvarianceEvidence struct {
	RecordedAt                 time.Time                     `json:"recorded_at"`
	PreUpgradeRuntimeIdentity  HermesRuntimeIdentityEvidence `json:"pre_upgrade_runtime_identity"`
	PostUpgradeRuntimeIdentity HermesRuntimeIdentityEvidence `json:"post_upgrade_runtime_identity"`
	BinaryUnchanged            bool                          `json:"binary_unchanged"`
	PreUpgradeConfigSHA256     string                        `json:"pre_upgrade_config_sha256"`
	PostUpgradeConfigSHA256    string                        `json:"post_upgrade_config_sha256"`
	ConfigUnchanged            bool                          `json:"config_unchanged"`
	CompatMode                 map[string]string             `json:"compat_mode"`
	Validated                  bool                          `json:"validated"`
	Error                      string                        `json:"error,omitempty"`
}

func buildHermesUpgradeInvarianceEvidence(
	preUpgradeIdentity HermesRuntimeIdentityEvidence,
	postUpgradeIdentity HermesRuntimeIdentityEvidence,
	preUpgradeConfig []byte,
	postUpgradeConfig []byte,
	chainIDs []string,
) (hermesUpgradeInvarianceEvidence, error) {
	evidence := hermesUpgradeInvarianceEvidence{
		RecordedAt:                 time.Now().UTC(),
		PreUpgradeRuntimeIdentity:  preUpgradeIdentity,
		PostUpgradeRuntimeIdentity: postUpgradeIdentity,
		BinaryUnchanged:            preUpgradeIdentity == postUpgradeIdentity,
		PreUpgradeConfigSHA256:     checksumBytes(preUpgradeConfig),
		PostUpgradeConfigSHA256:    checksumBytes(postUpgradeConfig),
		ConfigUnchanged:            len(preUpgradeConfig) != 0 && bytes.Equal(preUpgradeConfig, postUpgradeConfig),
	}

	var validationErrors []error
	if !evidence.BinaryUnchanged {
		validationErrors = append(validationErrors, errors.New("Hermes runtime binary identity changed during the Panacea upgrade"))
	}
	if !evidence.ConfigUnchanged {
		validationErrors = append(validationErrors, errors.New("Hermes config profile changed during the Panacea upgrade"))
	}
	compatMode, compatErr := parseHermesCompatMode(postUpgradeConfig, chainIDs)
	evidence.CompatMode = compatMode
	if compatErr != nil {
		validationErrors = append(validationErrors, compatErr)
	} else {
		for _, chainID := range chainIDs {
			if mode := compatMode[chainID]; mode != hermesExpectedCompatibilityMode {
				validationErrors = append(validationErrors, fmt.Errorf(
					"Hermes live compat_mode[%q] = %q, want %s",
					chainID,
					mode,
					hermesExpectedCompatibilityMode,
				))
			}
		}
	}
	joined := errors.Join(validationErrors...)
	evidence.Validated = joined == nil
	evidence.Error = errorString(joined)
	return evidence, joined
}

func parseHermesCompatMode(config []byte, chainIDs []string) (map[string]string, error) {
	if len(config) == 0 {
		return nil, errors.New("Hermes live config is empty")
	}
	tree, err := toml.LoadBytes(config)
	if err != nil {
		return nil, fmt.Errorf("parse live Hermes config: %w", err)
	}
	chains, ok := tree.Get("chains").([]*toml.Tree)
	if !ok || len(chains) == 0 {
		return nil, errors.New("live Hermes config has no chain tables")
	}
	expected := make(map[string]struct{}, len(chainIDs))
	for _, chainID := range chainIDs {
		chainID = strings.TrimSpace(chainID)
		if chainID == "" {
			return nil, errors.New("Hermes compat_mode chain ID is required")
		}
		if _, duplicate := expected[chainID]; duplicate {
			return nil, fmt.Errorf("Hermes compat_mode chain ID %q is duplicated", chainID)
		}
		expected[chainID] = struct{}{}
	}
	if len(expected) == 0 {
		return nil, errors.New("Hermes compat_mode requires at least one chain ID")
	}

	result := make(map[string]string, len(expected))
	for _, chain := range chains {
		chainID, ok := chain.Get("id").(string)
		if !ok {
			continue
		}
		chainID = strings.TrimSpace(chainID)
		if _, wanted := expected[chainID]; !wanted {
			continue
		}
		if _, duplicate := result[chainID]; duplicate {
			return result, fmt.Errorf("live Hermes config duplicates chain %q", chainID)
		}
		mode, ok := chain.Get("compat_mode").(string)
		if !ok || strings.TrimSpace(mode) == "" {
			return result, fmt.Errorf("live Hermes config chain %q has no compat_mode", chainID)
		}
		result[chainID] = strings.TrimSpace(mode)
	}
	for chainID := range expected {
		if _, present := result[chainID]; !present {
			return result, fmt.Errorf("live Hermes config is missing compat_mode for chain %q", chainID)
		}
	}
	return result, nil
}

func checksumBytes(contents []byte) string {
	checksum := sha256.Sum256(contents)
	return hex.EncodeToString(checksum[:])
}

func captureIBCChainBinaryEvidence(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	phase string,
	contract IBCBinaryVersionContract,
	store *artifactStore,
	artifactRoot string,
) (IBCChainBinaryEvidence, error) {
	evidence := IBCChainBinaryEvidence{Phase: phase, Contract: contract}
	if ctx == nil || chain == nil || store == nil {
		err := errors.New("IBC binary capture requires context, chain, and artifact store")
		evidence.Error = err.Error()
		return evidence, err
	}
	evidence.ChainID = chain.Config().ChainID
	nodes := chain.Nodes()
	if len(nodes) == 0 {
		err := errors.New("IBC binary capture chain has no nodes")
		evidence.Error = err.Error()
		_ = store.writeJSON(filepath.Join(artifactRoot, "identity.json"), evidence)
		return evidence, err
	}

	var captureErrors []error
	for index, node := range nodes {
		if node == nil {
			captureErrors = append(captureErrors, fmt.Errorf("IBC binary node %d is nil", index))
			continue
		}
		nodeRoot := filepath.Join(artifactRoot, "nodes", node.Name())
		nodeEvidence := IBCBinaryNodeEvidence{Name: node.Name()}

		versionStdout, versionStderr, versionErr := node.ExecBin(ctx, "version", "--long")
		if writeErr := store.write(filepath.Join(nodeRoot, "version-long.txt"), versionStdout); writeErr != nil {
			captureErrors = append(captureErrors, writeErr)
		}
		if len(versionStderr) != 0 {
			if writeErr := store.write(filepath.Join(nodeRoot, "version-long.stderr.txt"), versionStderr); writeErr != nil {
				captureErrors = append(captureErrors, writeErr)
			}
		}
		if versionErr != nil {
			captureErrors = append(captureErrors, fmt.Errorf(
				"read %s version --long on %s: %w: %s",
				contract.AppName,
				node.Name(),
				versionErr,
				strings.TrimSpace(string(versionStderr)),
			))
		} else {
			identity, parseErr := parseAndValidateIBCBinaryVersionLong(versionStdout, contract)
			nodeEvidence.Version = identity
			if parseErr != nil {
				captureErrors = append(captureErrors, fmt.Errorf("validate %s version on %s: %w", contract.AppName, node.Name(), parseErr))
			}
		}

		checksum, checksumErr := checksumIBCContainerBinary(ctx, node, contract.BinaryPath)
		checksumStdout := []byte(checksum + "  " + contract.BinaryPath + "\n")
		if writeErr := store.write(filepath.Join(nodeRoot, "binary-sha256.txt"), checksumStdout); writeErr != nil {
			captureErrors = append(captureErrors, writeErr)
		}
		if checksumErr != nil {
			captureErrors = append(captureErrors, fmt.Errorf(
				"checksum %s on %s: %w",
				contract.AppName,
				node.Name(),
				checksumErr,
			))
		} else {
			nodeEvidence.BinarySHA256 = checksum
			nodeEvidence.BinaryPath = contract.BinaryPath
		}
		evidence.Nodes = append(evidence.Nodes, nodeEvidence)
	}

	evidence.Validated = true
	if err := evidence.Validate(); err != nil {
		captureErrors = append(captureErrors, err)
	}
	joined := errors.Join(captureErrors...)
	if joined != nil {
		evidence.Validated = false
		evidence.Error = joined.Error()
	}
	writeErr := store.writeJSON(filepath.Join(artifactRoot, "identity.json"), evidence)
	if writeErr != nil {
		joined = errors.Join(joined, writeErr)
	}
	return evidence, joined
}

// checksumIBCContainerBinary reads the immutable executable through Docker's
// archive API. This works for the pinned distroless Osmosis image, which
// intentionally contains neither a shell nor sha256sum.
func checksumIBCContainerBinary(
	ctx context.Context,
	node *cosmos.ChainNode,
	binaryPath string,
) (string, error) {
	if ctx == nil || node == nil || node.DockerClient == nil || strings.TrimSpace(node.ContainerID()) == "" {
		return "", errors.New("IBC container binary checksum runtime is incomplete")
	}
	if !filepath.IsAbs(binaryPath) {
		return "", fmt.Errorf("IBC container binary path %q is not absolute", binaryPath)
	}
	archive, _, err := node.DockerClient.CopyFromContainer(ctx, node.ContainerID(), binaryPath)
	if err != nil {
		return "", fmt.Errorf("copy %s from container %s: %w", binaryPath, node.Name(), err)
	}
	defer archive.Close()
	return checksumIBCContainerBinaryArchive(archive, binaryPath)
}

func checksumIBCContainerBinaryArchive(archive io.Reader, binaryPath string) (string, error) {
	if archive == nil || !filepath.IsAbs(binaryPath) {
		return "", errors.New("IBC container binary archive and absolute path are required")
	}
	reader := tar.NewReader(archive)
	found := false
	var checksum string
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read Docker archive for %s: %w", binaryPath, err)
		}
		if header.FileInfo().IsDir() {
			continue
		}
		if !header.FileInfo().Mode().IsRegular() || filepath.Base(header.Name) != filepath.Base(binaryPath) {
			return "", fmt.Errorf("Docker archive entry %q is not the expected regular binary %s", header.Name, binaryPath)
		}
		if found {
			return "", fmt.Errorf("Docker archive for %s contains multiple binary entries", binaryPath)
		}
		if header.Size <= 0 || header.Size > ibcBinaryChecksumMaxBytes {
			return "", fmt.Errorf("container binary %s size %d is outside (0, %d]", binaryPath, header.Size, ibcBinaryChecksumMaxBytes)
		}
		hasher := sha256.New()
		written, err := io.Copy(hasher, io.LimitReader(reader, ibcBinaryChecksumMaxBytes+1))
		if err != nil {
			return "", fmt.Errorf("hash container binary %s: %w", binaryPath, err)
		}
		if written != header.Size {
			return "", fmt.Errorf("container binary %s archive size = %d, header declared %d", binaryPath, written, header.Size)
		}
		checksum = hex.EncodeToString(hasher.Sum(nil))
		found = true
	}
	if !found {
		return "", fmt.Errorf("Docker archive for %s contains no binary", binaryPath)
	}
	return checksum, nil
}

func captureIBCGenesisChecksumSnapshot(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	phase string,
	store *artifactStore,
	artifactRoot string,
	expectedSHA256 string,
) (IBCGenesisChecksumSnapshot, error) {
	snapshot := IBCGenesisChecksumSnapshot{Phase: phase}
	if ctx == nil || chain == nil || store == nil {
		return snapshot, errors.New("IBC genesis capture requires context, chain, and artifact store")
	}
	nodes := chain.Nodes()
	if len(nodes) == 0 {
		return snapshot, errors.New("IBC genesis capture chain has no nodes")
	}
	var captureErrors []error
	for index, node := range nodes {
		if node == nil {
			captureErrors = append(captureErrors, fmt.Errorf("IBC genesis node %d is nil", index))
			continue
		}
		contents, err := node.ReadFile(ctx, "config/genesis.json")
		if err != nil {
			captureErrors = append(captureErrors, fmt.Errorf("read genesis from %s: %w", node.Name(), err))
			continue
		}
		if phase == "pre-upgrade" && index == 0 {
			if err := store.write(filepath.Join(artifactRoot, "genesis.json"), contents); err != nil {
				captureErrors = append(captureErrors, err)
			}
		}
		checksum := checksumGenesis(node.Name(), contents)
		snapshot.Nodes = append(snapshot.Nodes, checksum)
		if snapshot.Common == "" {
			snapshot.Common = checksum.SHA256
		}
	}
	if err := snapshot.Validate(); err != nil {
		captureErrors = append(captureErrors, err)
	}
	if expectedSHA256 != "" && snapshot.Common != expectedSHA256 {
		captureErrors = append(captureErrors, fmt.Errorf(
			"%s genesis checksum = %s, want immutable %s",
			phase,
			snapshot.Common,
			expectedSHA256,
		))
	}
	writeErr := store.writeJSON(filepath.Join(artifactRoot, "genesis-checksums-"+phase+".json"), snapshot)
	return snapshot, errors.Join(append(captureErrors, writeErr)...)
}

func (n *IBCTopology) captureHermesUpgradeInvariance(ctx context.Context) (hermesUpgradeInvarianceEvidence, error) {
	var zero hermesUpgradeInvarianceEvidence
	if ctx == nil || n == nil || n.artifacts == nil || n.artifacts.base == nil || n.hermes == nil {
		return zero, errors.New("Hermes upgrade invariance runtime is incomplete")
	}

	var captureErrors []error
	baselinePath, err := n.artifacts.base.safePath(hermesPostChannelConfigArtifactPath)
	if err != nil {
		captureErrors = append(captureErrors, err)
	}
	var preUpgradeConfig []byte
	if err == nil {
		preUpgradeConfig, err = os.ReadFile(baselinePath)
		if err != nil {
			captureErrors = append(captureErrors, fmt.Errorf("read pre-upgrade Hermes config artifact: %w", err))
		}
	}

	postUpgradeConfig, configErr := n.hermes.ReadFileFromHomeDir(ctx, ".hermes/config.toml")
	if configErr != nil {
		captureErrors = append(captureErrors, fmt.Errorf("read post-upgrade live Hermes config: %w", configErr))
	} else if writeErr := n.artifacts.base.write(hermesPostUpgradeConfigArtifactPath, postUpgradeConfig); writeErr != nil {
		captureErrors = append(captureErrors, writeErr)
	}

	version, versionErr := n.artifacts.execHermesEvidence(
		ctx,
		n.hermes,
		n.execReporter,
		"version-post-upgrade",
		[]string{"hermes", "version"},
	)
	if versionErr != nil {
		captureErrors = append(captureErrors, versionErr)
	}
	checksum, checksumErr := n.artifacts.execHermesEvidence(
		ctx,
		n.hermes,
		n.execReporter,
		"binary-sha256-post-upgrade",
		[]string{"sh", "-c", `set -eu; hermes_path="$(command -v hermes)"; sha256sum "$hermes_path"`},
	)
	if checksumErr != nil {
		captureErrors = append(captureErrors, checksumErr)
	}
	var postUpgradeIdentity HermesRuntimeIdentityEvidence
	if versionErr == nil && checksumErr == nil {
		postUpgradeIdentity, err = validateHermesRuntimeIdentity(version, checksum, PinnedIBCProvenance())
		if err != nil {
			captureErrors = append(captureErrors, err)
		}
	}

	chainIDs := []string{n.Panacea.Config().ChainID, n.Osmosis.Config().ChainID}
	evidence, validationErr := buildHermesUpgradeInvarianceEvidence(
		n.hermesRuntime,
		postUpgradeIdentity,
		preUpgradeConfig,
		postUpgradeConfig,
		chainIDs,
	)
	joined := errors.Join(append(captureErrors, validationErr)...)
	if joined != nil {
		evidence.Validated = false
		evidence.Error = joined.Error()
	}
	writeErr := n.artifacts.base.writeJSON(hermesUpgradeInvarianceArtifactPath, evidence)
	return evidence, errors.Join(joined, writeErr)
}

func (n *IBCTopology) recordIBCCompatibilityMatrix(ctx context.Context) error {
	if n == nil || n.artifacts == nil || n.artifacts.base == nil || n.channel == nil {
		return errors.New("IBC compatibility matrix runtime is incomplete")
	}
	hermesInvariance, hermesInvarianceErr := n.captureHermesUpgradeInvariance(ctx)
	matrix := IBCCompatibilityMatrix{
		SchemaVersion: ibcCompatibilityMatrixSchema,
		GeneratedAt:   time.Now().UTC(),
		Panacea: IBCBinaryUpgradeMatrix{
			PreUpgrade:  n.panaceaPreUpgradeBinary,
			PostUpgrade: n.panaceaPostUpgradeBinary,
		},
		Osmosis: IBCOsmosisCompatibilityMatrix{
			SourceContract: pinnedOsmosisSourceContract(),
			PreUpgrade:     n.osmosisPreUpgradeBinary,
			PostUpgrade:    n.osmosisPostUpgradeBinary,
			Genesis:        n.osmosisGenesisImmutability,
		},
		Channel: IBCCompatibilityChannelContract{
			PanaceaChannelID: n.channel.Panacea.ChannelID,
			OsmosisChannelID: n.channel.Osmosis.ChannelID,
			PortID:           n.channel.Panacea.PortID,
			State:            n.channel.Panacea.ChannelState,
			Ordering:         n.channel.Panacea.Ordering,
			Version:          n.channel.Panacea.Version,
		},
		Middleware: expectedOsmosisMiddlewareEvidence(),
		Hermes: IBCHermesCompatibilityMatrix{
			RuntimeIdentity: hermesInvariance.PostUpgradeRuntimeIdentity,
			CompatMode:      hermesInvariance.CompatMode,
		},
		Validated: hermesInvarianceErr == nil,
		Error:     errorString(hermesInvarianceErr),
	}
	validationErr := matrix.Validate()
	if validationErr != nil {
		matrix.Validated = false
		matrix.Error = validationErr.Error()
	}
	writeErr := n.artifacts.base.writeJSON(IBCCompatibilityMatrixArtifactPath, matrix)
	return errors.Join(hermesInvarianceErr, validationErr, writeErr)
}
